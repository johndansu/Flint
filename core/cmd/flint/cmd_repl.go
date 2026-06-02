package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"flint/core/internal/ai"
	"flint/core/internal/config"
	"flint/core/internal/store"
)

const maxExchanges = 20

// replLockPath returns the sentinel file the daemon checks to suppress
// passive observations while a REPL session is active.
func replLockPath() string {
	return filepath.Join(config.FlintDir(), "repl.lock")
}

func cmdRepl(_ []string) error {
	cfg, err := config.Load(workspaceRoot())
	if err != nil {
		return err
	}

	if cfg.APIKey == "" {
		return fmt.Errorf("no API key — run 'flint init' to configure")
	}

	db, err := store.Open(config.FlintDir())
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer db.Close()

	root := workspaceRoot()

	// Write a lock file so the daemon suppresses passive observations while
	// the developer is in an intentional REPL conversation.
	lockPath := replLockPath()
	_ = os.WriteFile(lockPath, []byte("1"), 0o600)
	defer os.Remove(lockPath)

	fmt.Printf("\nLoading context for %s...", filepath.Base(root))
	rc := loadReplContext(db, root)
	system := buildReplSystem(root, rc)
	fmt.Println(" ready.")

	client := ai.New(cfg.APIKey, cfg.Model)

	// Session header
	fmt.Printf("\nFlint / %s\n", filepath.Base(root))
	fmt.Println(separator())
	fmt.Printf("  Stack:     %s\n", rc.summary)
	fmt.Printf("  Awareness: %s\n", cfg.Awareness)
	if rc.gitBranch != "" {
		fmt.Printf("  Branch:    %s\n", rc.gitBranch)
	}
	fmt.Println("  Type 'exit' to close  ·  Ctrl+C to quit")
	fmt.Println(separator())
	fmt.Println()

	// Ctrl+C exits cleanly without a stack trace
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		os.Remove(lockPath)
		fmt.Println("\nSession closed.")
		os.Exit(0)
	}()

	var messages []ai.Message

	for {
		fmt.Print("> ")
		line := strings.TrimSpace(readLine())

		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" || line == ":q" {
			fmt.Println("\nSession closed.")
			return nil
		}

		messages = append(messages, ai.Message{Role: "user", Content: line})

		// Hard context window: drop oldest user+assistant exchange when over limit
		// so the session never hits the API token ceiling.
		for len(messages) > maxExchanges*2 {
			messages = messages[2:]
		}

		fmt.Println()
		var reply strings.Builder
		_, err := client.Stream(context.Background(), system, messages,
			func(delta string) {
				fmt.Print(delta)
				reply.WriteString(delta)
			},
		)
		fmt.Print("\n\n")

		if err != nil {
			fmt.Printf("Error: %v\n\n", err)
			messages = messages[:len(messages)-1] // drop failed user message
			continue
		}

		messages = append(messages, ai.Message{Role: "assistant", Content: reply.String()})
	}
}

// ---------------------------------------------------------------------------
// Context
// ---------------------------------------------------------------------------

type replContext struct {
	// From scan baselines
	projectName string
	totalFiles  string
	goFiles     string
	tsFiles     string
	pyFiles     string
	gitRemote   string
	scannedAt   string

	// Live from disk / git
	gitBranch      string
	gitLog         string // last N commits, one per line
	dirTree        string // top-level structure
	structFiles    map[string]string // filename → content
	readmeExcerpt  string
	docExcerpts    map[string]string // ARCHITECTURE.md, DESIGN.md, CLAUDE.md etc.

	// From store
	recentObs    []string // observation texts
	recentErrors []string // error messages

	summary string // one-line description for the header
}

func loadReplContext(db *store.DB, root string) replContext {
	get := func(key string) string {
		v, _ := db.GetBaseline(key)
		return v
	}

	rc := replContext{
		projectName: get("project.name"),
		totalFiles:  get("scan.files"),
		goFiles:     get("scan.go_files"),
		tsFiles:     get("scan.ts_files"),
		pyFiles:     get("scan.python_files"),
		gitRemote:   get("git.remote"),
		scannedAt:   get("scan.scanned_at"),
		structFiles: make(map[string]string),
		docExcerpts: make(map[string]string),
	}

	if rc.projectName == "" {
		rc.projectName = filepath.Base(root)
	}

	// ── Live: git branch
	if out, err := exec.Command("git", "-C", root, "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
		rc.gitBranch = strings.TrimSpace(string(out))
	}

	// ── Live: recent git commits
	if out, err := exec.Command(
		"git", "-C", root, "log", "--oneline", "-12",
		"--format=%h  %an  %ar  %s",
	).Output(); err == nil {
		rc.gitLog = strings.TrimSpace(string(out))
	}

	// ── Live: top-level directory structure (depth 2, skipping noise)
	rc.dirTree = buildDirTree(root, 2)

	// ── Live: key structural files
	structuralFiles := []string{
		"go.mod", "package.json", "Cargo.toml",
		"pyproject.toml", "requirements.txt",
		"Makefile", "docker-compose.yml", ".env.example",
	}
	for _, name := range structuralFiles {
		if content := readIfSmall(filepath.Join(root, name), 3000); content != "" {
			rc.structFiles[name] = content
		}
	}

	// ── Live: README
	for _, name := range []string{"README.md", "README.rst", "README"} {
		if content := readIfSmall(filepath.Join(root, name), 3000); content != "" {
			rc.readmeExcerpt = content
			break
		}
	}

	// ── Live: architecture / design docs
	docFiles := []string{
		"ARCHITECTURE.md", "DESIGN.md", "CLAUDE.md",
		"docs/ARCHITECTURE.md", "docs/architecture.md",
		"docs/DESIGN.md", "docs/overview.md",
	}
	for _, name := range docFiles {
		if content := readIfSmall(filepath.Join(root, name), 2000); content != "" {
			rc.docExcerpts[name] = content
		}
	}

	// ── Store: recent observations (text only — the AI knows what Flint has flagged)
	if rows, err := db.Raw().Query(
		`SELECT text FROM observations ORDER BY timestamp DESC LIMIT 5`,
	); err == nil {
		defer rows.Close()
		for rows.Next() {
			var text string
			if rows.Scan(&text) == nil {
				rc.recentObs = append(rc.recentObs, text)
			}
		}
	}

	// ── Store: recent errors
	if errs, err := db.RecentErrors(5); err == nil {
		for _, e := range errs {
			msg := e.Source + ": " + e.Message
			if e.FilePath != "" {
				msg += " (" + e.FilePath
				if e.LineNumber > 0 {
					msg += fmt.Sprintf(":%d", e.LineNumber)
				}
				msg += ")"
			}
			rc.recentErrors = append(rc.recentErrors, msg)
		}
	}

	// ── Header summary
	var parts []string
	if rc.totalFiles != "" && rc.totalFiles != "0" {
		parts = append(parts, rc.totalFiles+" files")
	}
	for _, pair := range [][2]string{
		{rc.goFiles, "Go"}, {rc.tsFiles, "TypeScript"}, {rc.pyFiles, "Python"},
	} {
		if pair[0] != "" && pair[0] != "0" {
			parts = append(parts, pair[0]+" "+pair[1])
		}
	}
	if len(parts) > 0 {
		rc.summary = strings.Join(parts, ", ")
	} else {
		rc.summary = "no scan baseline — run 'flint scan' for richer context"
	}

	return rc
}

// ---------------------------------------------------------------------------
// System prompt
// ---------------------------------------------------------------------------

func buildReplSystem(root string, rc replContext) string {
	w := &strings.Builder{}

	section := func(title string) {
		fmt.Fprintf(w, "\n## %s\n", title)
	}
	block := func(label, content string) {
		if strings.TrimSpace(content) == "" {
			return
		}
		fmt.Fprintf(w, "\n--- %s ---\n%s\n", label, strings.TrimSpace(content))
	}

	// ── Role
	section("Role")
	fmt.Fprintf(w, `You are Flint — a senior developer who has studied the codebase at %s and knows it well.
You answer the way a senior dev answers a teammate: specific, direct, grounded in the actual code, never generic.
You are NOT a general-purpose assistant. Every answer must be rooted in this project's actual context.
If something isn't in your context, say "I'd need to read [file/area]" rather than guessing.`, root)

	// ── Project Identity
	section("Project Identity")
	fmt.Fprintf(w, "Name:      %s\n", rc.projectName)
	fmt.Fprintf(w, "Root:      %s\n", root)
	if rc.gitRemote != "" {
		fmt.Fprintf(w, "Repo:      %s\n", rc.gitRemote)
	}
	if rc.gitBranch != "" {
		fmt.Fprintf(w, "Branch:    %s\n", rc.gitBranch)
	}
	if rc.scannedAt != "" {
		fmt.Fprintf(w, "Scanned:   %s\n", rc.scannedAt)
	}

	var langParts []string
	for _, pair := range [][2]string{
		{rc.goFiles, "Go"}, {rc.tsFiles, "TypeScript"}, {rc.pyFiles, "Python"},
	} {
		if pair[0] != "" && pair[0] != "0" {
			langParts = append(langParts, pair[0]+" "+pair[1])
		}
	}
	if len(langParts) > 0 {
		fmt.Fprintf(w, "Stack:     %s (%s total files)\n", strings.Join(langParts, ", "), rc.totalFiles)
	} else if rc.totalFiles != "" {
		fmt.Fprintf(w, "Files:     %s total\n", rc.totalFiles)
	}

	// ── Project Structure
	if rc.dirTree != "" {
		section("Project Structure")
		w.WriteString(rc.dirTree)
	}

	// ── Structural files (go.mod, package.json, etc.)
	if len(rc.structFiles) > 0 {
		section("Build & Dependency Files")
		for _, name := range []string{
			"go.mod", "package.json", "Cargo.toml",
			"pyproject.toml", "requirements.txt",
			"Makefile", "docker-compose.yml", ".env.example",
		} {
			block(name, rc.structFiles[name])
		}
	}

	// ── README
	if rc.readmeExcerpt != "" {
		section("README")
		w.WriteString(rc.readmeExcerpt)
	}

	// ── Architecture / design docs
	if len(rc.docExcerpts) > 0 {
		section("Documentation")
		for name, content := range rc.docExcerpts {
			block(name, content)
		}
	}

	// ── Git history
	if rc.gitLog != "" {
		section("Recent Git History")
		w.WriteString(rc.gitLog)
	}

	// ── What Flint has already flagged
	if len(rc.recentObs) > 0 {
		section("Recent Flint Observations (what the daemon has flagged)")
		for i, obs := range rc.recentObs {
			fmt.Fprintf(w, "%d. %s\n", i+1, obs)
		}
	}

	// ── Recent errors in the log
	if len(rc.recentErrors) > 0 {
		section("Recent Errors (from flint error log)")
		for i, e := range rc.recentErrors {
			fmt.Fprintf(w, "%d. %s\n", i+1, e)
		}
	}

	// ── Session rules
	section("Session Rules")
	w.WriteString(`- Answer about THIS project specifically. Reference actual files, functions, and patterns you see above.
- Be concise: 3–5 sentences by default. Expand only when precision requires it.
- When asked about architecture: explain the WHY (constraints, trade-offs), not just the WHAT.
- When asked about a bug: reason through root cause step by step before suggesting a fix.
- When asked "what should I work on next": look at git history, errors, and observations above for real signals.
- When asked to review or explain code: be honest about risks, debt, and assumptions — don't flatter.
- Never invent file paths, function names, or API endpoints that aren't in the context above.
- If the answer requires reading a specific file you haven't seen: say so and tell the developer which file to paste in.
- This session is cleared when it closes. Nothing persists.`)

	return w.String()
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// buildDirTree returns a readable directory tree up to maxDepth levels deep,
// skipping noisy directories like node_modules, .git, vendor, dist, build.
func buildDirTree(root string, maxDepth int) string {
	skip := map[string]bool{
		"node_modules": true, ".git": true, "vendor": true,
		"dist": true, "build": true, "out": true, ".cache": true,
		".next": true, "__pycache__": true, ".venv": true,
	}

	var sb strings.Builder
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || path == root {
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		depth := strings.Count(rel, string(filepath.Separator))

		if d.IsDir() {
			if skip[d.Name()] {
				return filepath.SkipDir
			}
			if depth >= maxDepth {
				return filepath.SkipDir
			}
		} else if depth >= maxDepth {
			return nil
		}

		indent := strings.Repeat("  ", depth)
		name := d.Name()
		if d.IsDir() {
			name += "/"
		}
		fmt.Fprintf(&sb, "%s%s\n", indent, name)
		return nil
	})
	return sb.String()
}

// readIfSmall reads a file and returns its content when under maxBytes.
// Returns empty string if the file does not exist or exceeds the limit.
func readIfSmall(path string, maxBytes int) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	if len(b) > maxBytes {
		return string(b[:maxBytes]) + "\n... (truncated)"
	}
	return string(b)
}
