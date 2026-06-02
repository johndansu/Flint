package daemon

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"flint/core/internal/config"
	"flint/core/internal/ipc"
	"flint/core/internal/store"
	schema "flint/core/pkg/schema"
)

const (
	burnoutThreshold      = 3 * time.Hour
	burnoutCooldown       = 24 * time.Hour
	spiralWindow          = 30 * time.Minute
	spiralMinEdits        = 10
	spiralCooldown        = 30 * time.Minute
	loneWolfDays          = 3
	loneWolfCooldown      = 7 * 24 * time.Hour
	loneWolfCheckInterval = time.Hour
)

// HumanLayer detects three human-factor patterns and fires kind:"human" observations:
//
//  1. Burnout          — coding for 3+ hours without a 30-min break
//  2. Debugging spiral — same file edited 10+ times in 30 minutes
//  3. Lone wolf        — no git commits in the last 3 days
type HumanLayer struct {
	server        *ipc.Server
	db            *store.DB
	workspaceRoot string
	thresh        config.Thresholds

	mu           sync.Mutex
	sessionStart time.Time

	// Debugging spiral: per-file edit timestamps within the spiral window
	fileEdits   map[string][]time.Time
	spiralFired map[string]time.Time // last fire time per file

	// Rate limiters
	lastBurnoutFired  time.Time
	lastLoneWolfFired time.Time
}

// NewHumanLayer creates a HumanLayer. Start the lone-wolf background check
// by calling RunLoneWolfCheck after creation.
func NewHumanLayer(server *ipc.Server, db *store.DB, workspaceRoot string, thresh config.Thresholds) *HumanLayer {
	return &HumanLayer{
		server:        server,
		db:            db,
		workspaceRoot: workspaceRoot,
		thresh:        thresh,
		sessionStart:  time.Now(),
		fileEdits:     make(map[string][]time.Time),
		spiralFired:   make(map[string]time.Time),
	}
}

// OnNewSession resets the burnout clock when a 30-min inactivity gap starts a
// new session — the developer has taken a real break.
func (h *HumanLayer) OnNewSession() {
	h.mu.Lock()
	h.sessionStart = time.Now()
	h.mu.Unlock()
}

// OnFileEvent tracks edits per file and checks for debugging spirals.
// Called on every file save event.
func (h *HumanLayer) OnFileEvent(event FileEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := event.Time
	path := event.Path
	cutoff := now.Add(-h.thresh.SpiralWindow())

	// Append and trim to the rolling window
	edits := append(h.fileEdits[path], now)
	fresh := edits[:0]
	for _, t := range edits {
		if t.After(cutoff) {
			fresh = append(fresh, t)
		}
	}
	h.fileEdits[path] = fresh

	// Spiral threshold: N+ edits to the same file in the window
	if len(fresh) < h.thresh.SpiralEditsCount() {
		return
	}
	if time.Since(h.spiralFired[path]) < spiralCooldown {
		return
	}
	h.spiralFired[path] = now

	editCount := len(fresh)
	go h.fireSpiral(path, editCount)
}

// CheckBurnout checks whether the current session has exceeded the burnout
// threshold. Call this from the classifier's periodic update cycle.
func (h *HumanLayer) CheckBurnout() {
	h.mu.Lock()
	duration := time.Since(h.sessionStart)
	lastFired := h.lastBurnoutFired
	h.mu.Unlock()

	if duration < h.thresh.BurnoutDuration() {
		return
	}
	if time.Since(lastFired) < burnoutCooldown {
		return
	}

	h.mu.Lock()
	h.lastBurnoutFired = time.Now()
	hours := int(duration.Hours())
	h.mu.Unlock()

	go h.fireBurnout(hours)
}

// RunLoneWolfCheck starts a background goroutine that checks for lone-wolf
// patterns once per hour. Call once after NewHumanLayer.
func (h *HumanLayer) RunLoneWolfCheck(ctx context.Context) {
	go func() {
		// Check immediately on start so the first run isn't delayed an hour
		h.checkLoneWolf()

		ticker := time.NewTicker(loneWolfCheckInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				h.checkLoneWolf()
			}
		}
	}()
}

func (h *HumanLayer) checkLoneWolf() {
	h.mu.Lock()
	if time.Since(h.lastLoneWolfFired) < loneWolfCooldown {
		h.mu.Unlock()
		return
	}
	h.mu.Unlock()

	days := h.thresh.LoneWolfDays
	if days <= 0 {
		days = 3
	}
	since := fmt.Sprintf("%d days ago", days)
	out, err := exec.Command(
		"git", "-C", h.workspaceRoot,
		"log", "--oneline", "--since="+since,
	).Output()
	if err != nil {
		return // not a git repo or git not installed
	}
	if strings.TrimSpace(string(out)) != "" {
		return // recent commits exist — not lone wolf
	}

	h.mu.Lock()
	h.lastLoneWolfFired = time.Now()
	h.mu.Unlock()

	go h.fireLoneWolf()
}

// ---------------------------------------------------------------------------
// Observation builders
// ---------------------------------------------------------------------------

func (h *HumanLayer) fireSpiral(filePath string, editCount int) {
	text := fmt.Sprintf(
		"You've edited %s %d times in the last 30 minutes — that's a debugging spiral. "+
			"Consider stepping away for 10 minutes, explaining the problem out loud, or asking a teammate for a fresh eye.",
		filepath.Base(filePath), editCount,
	)
	if h.server != nil {
		h.server.BroadcastSessionState(schema.SessionState{
			State:       schema.SessionStateStateDebuggingSpiral,
			SessionType: schema.SessionStateSessionTypeHuman,
			Since:       time.Now().UTC().Format(time.RFC3339),
		})
	}
	h.broadcast(schema.ObservationKindHuman, "complexity", 0.90, text, &filePath)
	log.Printf("humanlayer: debugging spiral on %s (%d edits in 30 min)", filepath.Base(filePath), editCount)
}

func (h *HumanLayer) fireBurnout(hours int) {
	text := fmt.Sprintf(
		"You've been coding for %d hours without a significant break. "+
			"Cognitive performance drops sharply after 90 minutes of focused work — "+
			"a 15-minute break now will likely save you 60 minutes of slower thinking later.",
		hours,
	)
	h.broadcast(schema.ObservationKindHuman, "other", 0.85, text, nil)
	log.Printf("humanlayer: burnout signal (%d hours continuous session)", hours)
}

func (h *HumanLayer) fireLoneWolf() {
	text := fmt.Sprintf(
		"No commits to this repository in %d days. "+
			"Long stretches without commits mean larger diffs, harder reviews, and higher merge-conflict risk. "+
			"Consider a WIP commit to checkpoint your work.",
		loneWolfDays,
	)
	h.broadcast(schema.ObservationKindHuman, "other", 0.80, text, nil)
	log.Printf("humanlayer: lone-wolf signal (no commits in %d days)", loneWolfDays)
}

func (h *HumanLayer) broadcast(kind, category string, confidence float64, text string, filePath *string) {
	obs := schema.Observation{
		Id:         uuid.New().String(),
		Kind:       kind,
		Category:   category,
		Confidence: confidence,
		Text:       text,
		FilePath:   filePath,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}
	if h.server != nil {
		h.server.BroadcastObservation(obs)
	}

	// Persist to store so it appears in flint status / fix
	_ = h.db.InsertObservation(store.Observation{
		ID:         obs.Id,
		Kind:       obs.Kind,
		Category:   obs.Category,
		Confidence: obs.Confidence,
		Text:       obs.Text,
		Timestamp:  time.Now().UTC(),
	})
}
