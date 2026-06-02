package daemon

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"flint/core/internal/ipc"
	"flint/core/internal/store"
	schema "flint/core/pkg/schema"
)

const (
	gitCheckInterval    = 30 * time.Minute
	staleBranchDays     = 3
	staleBranchCooldown = 3 * 24 * time.Hour
	uncommittedAge      = 2 * time.Hour
	uncommittedCooldown = 4 * time.Hour
	msgCooldown         = 7 * 24 * time.Hour // re-scan messages weekly
)

// poorMsgWords are commit subjects that signal a low-quality message.
var poorMsgWords = map[string]bool{
	"fix": true, "wip": true, "update": true, "done": true,
	"changes": true, "ok": true, "test": true, "stuff": true,
	"temp": true, "tmp": true, "work": true, "more": true,
	"misc": true, "cleanup": true, "minor": true,
}

// GitMonitor watches the workspace's git repo and fires observations for:
//
//  1. Stale feature branch — no commits in 3+ days
//  2. Poor commit message quality — vague one-word subjects
//  3. Long-lived uncommitted changes — dirty working tree for 2+ hours
type GitMonitor struct {
	server        *ipc.Server
	db            *store.DB
	workspaceRoot string

	mu                   sync.Mutex
	lastUncommittedFired time.Time
	firedMsgHashes       map[string]bool // commit hashes already flagged
}

// NewGitMonitor creates a GitMonitor. Call Run to start the background loop.
func NewGitMonitor(server *ipc.Server, db *store.DB, workspaceRoot string) *GitMonitor {
	// Seed fired message hashes from persisted baselines
	fired := make(map[string]bool)
	if v, _ := db.GetBaseline("git.fired_msg_hashes"); v != "" {
		for _, h := range strings.Split(v, ",") {
			if h = strings.TrimSpace(h); h != "" {
				fired[h] = true
			}
		}
	}
	return &GitMonitor{
		server:         server,
		db:             db,
		workspaceRoot:  workspaceRoot,
		firedMsgHashes: fired,
	}
}

// Run starts the background git-health loop. Checks immediately, then every 30 min.
func (g *GitMonitor) Run(ctx context.Context) {
	go func() {
		g.check()

		ticker := time.NewTicker(gitCheckInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				g.check()
			}
		}
	}()
}

func (g *GitMonitor) check() {
	if !g.isGitRepo() {
		return
	}
	g.checkStaleBranch()
	g.checkPoorMessages()
	g.checkUncommitted()
}

// ---------------------------------------------------------------------------
// Check: stale branch
// ---------------------------------------------------------------------------

func (g *GitMonitor) checkStaleBranch() {
	branch := g.gitOutput("rev-parse", "--abbrev-ref", "HEAD")
	if branch == "" || branch == "HEAD" {
		return
	}
	// Skip trunk branches
	switch branch {
	case "main", "master", "develop", "dev", "trunk":
		return
	}

	// Last commit timestamp on this branch
	tsStr := g.gitOutput("log", "-1", "--format=%ct", branch)
	if tsStr == "" {
		return
	}
	var epochSec int64
	fmt.Sscanf(tsStr, "%d", &epochSec)
	lastCommit := time.Unix(epochSec, 0)
	age := time.Since(lastCommit)
	if age < staleBranchDays*24*time.Hour {
		return
	}

	// Dedup via baseline: key = "git.stale_branch.<branch>"
	key := "git.stale_branch." + branch
	if v, _ := g.db.GetBaseline(key); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err == nil && time.Since(t) < staleBranchCooldown {
			return
		}
	}
	_ = g.db.SetBaseline(key, time.Now().UTC().Format(time.RFC3339))

	days := int(age.Hours() / 24)
	text := fmt.Sprintf(
		"Branch '%s' has had no commits for %d days. Consider merging or cleaning it up before it diverges further.",
		branch, days,
	)
	g.fire(schema.ObservationKindTechnical, "git", 0.70, text, "")
	log.Printf("git: stale branch '%s' (%d days)", branch, days)
}

// ---------------------------------------------------------------------------
// Check: poor commit messages
// ---------------------------------------------------------------------------

func (g *GitMonitor) checkPoorMessages() {
	// "git log --format=<hash> <subject>" for last 20 commits
	out := g.gitOutput("log", "--format=%H\t%s", "-20")
	if out == "" {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	var newHashes []string
	for _, line := range strings.Split(out, "\n") {
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		hash, subject := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if hash == "" || g.firedMsgHashes[hash] {
			continue
		}
		if isPoorMessage(subject) {
			g.firedMsgHashes[hash] = true
			newHashes = append(newHashes, hash)
			text := fmt.Sprintf(
				"Commit %s has a vague message: %q. Future-you will thank present-you for a descriptive subject line.",
				hash[:7], subject,
			)
			go g.fire(schema.ObservationKindTaste, "git", 0.65, text, "")
			log.Printf("git: poor commit message %s %q", hash[:7], subject)
		}
	}

	if len(newHashes) > 0 {
		// Persist updated set
		all := make([]string, 0, len(g.firedMsgHashes))
		for h := range g.firedMsgHashes {
			all = append(all, h)
		}
		_ = g.db.SetBaseline("git.fired_msg_hashes", strings.Join(all, ","))
	}
}

func isPoorMessage(subject string) bool {
	// Very short
	if len(subject) < 8 {
		return true
	}
	// Exact match against known vague words (case-insensitive)
	lower := strings.ToLower(strings.TrimSpace(subject))
	if poorMsgWords[lower] {
		return true
	}
	// Starts with a poor word and nothing else meaningful follows
	for word := range poorMsgWords {
		if lower == word || lower == word+"s" || lower == word+"ed" || lower == word+"ing" {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Check: long-lived uncommitted changes
// ---------------------------------------------------------------------------

func (g *GitMonitor) checkUncommitted() {
	g.mu.Lock()
	lastFired := g.lastUncommittedFired
	g.mu.Unlock()

	if time.Since(lastFired) < uncommittedCooldown {
		return
	}

	// Get list of modified tracked files
	out := g.gitOutput("status", "--porcelain")
	if out == "" {
		return
	}

	var oldestMod time.Time
	var stalePaths []string

	for _, line := range strings.Split(out, "\n") {
		if len(line) < 4 {
			continue
		}
		xy := line[:2]
		// Only care about modifications to tracked files, not untracked (??)
		if xy == "??" || xy == "  " {
			continue
		}
		relPath := strings.TrimSpace(line[3:])
		// Handle renames: "old -> new"
		if idx := strings.Index(relPath, " -> "); idx != -1 {
			relPath = relPath[idx+4:]
		}
		absPath := filepath.Join(g.workspaceRoot, relPath)
		info, err := os.Stat(absPath)
		if err != nil {
			continue
		}
		modTime := info.ModTime()
		if time.Since(modTime) < uncommittedAge {
			continue // recently touched — not stale
		}
		stalePaths = append(stalePaths, relPath)
		if oldestMod.IsZero() || modTime.Before(oldestMod) {
			oldestMod = modTime
		}
	}

	if len(stalePaths) == 0 {
		return
	}

	g.mu.Lock()
	g.lastUncommittedFired = time.Now()
	g.mu.Unlock()

	hours := int(time.Since(oldestMod).Hours())
	preview := stalePaths[0]
	extra := ""
	if len(stalePaths) > 1 {
		extra = fmt.Sprintf(" (+%d more)", len(stalePaths)-1)
	}
	text := fmt.Sprintf(
		"%d uncommitted file(s) have been sitting for %d+ hours (%s%s). Consider committing or stashing.",
		len(stalePaths), hours, preview, extra,
	)
	g.fire(schema.ObservationKindHuman, "git", 0.72, text, "")
	log.Printf("git: %d uncommitted file(s) idle for %d+ hours", len(stalePaths), hours)
}

// ---------------------------------------------------------------------------
// Broadcast + persist
// ---------------------------------------------------------------------------

func (g *GitMonitor) fire(kind, category string, confidence float64, text, filePath string) {
	obs := schema.Observation{
		Id:         uuid.New().String(),
		Kind:       kind,
		Category:   category,
		Confidence: confidence,
		Text:       text,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}
	if filePath != "" {
		obs.FilePath = &filePath
	}

	g.server.BroadcastObservation(obs)

	fp := ""
	if obs.FilePath != nil {
		fp = *obs.FilePath
	}
	_ = g.db.InsertObservation(store.Observation{
		ID:         obs.Id,
		Kind:       obs.Kind,
		Category:   obs.Category,
		Confidence: obs.Confidence,
		Text:       obs.Text,
		FilePath:   fp,
		Timestamp:  time.Now().UTC(),
	})
}

// ---------------------------------------------------------------------------
// Git helpers
// ---------------------------------------------------------------------------

func (g *GitMonitor) isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = g.workspaceRoot
	return cmd.Run() == nil
}

func (g *GitMonitor) gitOutput(args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.workspaceRoot
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
