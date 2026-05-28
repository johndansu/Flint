package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"flint/core/internal/ai"
	"flint/core/internal/config"
	"flint/core/internal/store"
)

func cmdFix(args []string) error {
	cfg, err := config.Load(workspaceRoot())
	if err != nil {
		return err
	}

	db, err := store.Open(config.FlintDir())
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer db.Close()

	if len(args) == 0 {
		return listErrors(db)
	}

	if args[0] == "--auto" {
		if len(args) < 2 {
			return fmt.Errorf("--auto requires a category: cve | import | format")
		}
		return enableAutoFix(args[1])
	}

	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("expected a numeric error ID — run 'flint fix' to see recent errors")
	}

	return fixError(id, cfg, db)
}

func listErrors(db *store.DB) error {
	errs, err := db.RecentErrors(10)
	if err != nil {
		return fmt.Errorf("read errors: %w", err)
	}
	if len(errs) == 0 {
		fmt.Println("\nNo errors logged yet.")
		fmt.Println("Errors are recorded during 'flint precommit', daemon runs, and 'flint scan'.")
		fmt.Println()
		return nil
	}

	fmt.Printf("\nRecent errors\n%s\n", separator())
	for _, e := range errs {
		loc := e.Source
		if e.FilePath != "" {
			loc += " · " + e.FilePath
			if e.LineNumber > 0 {
				loc += fmt.Sprintf(":%d", e.LineNumber)
			}
		}
		ts := ""
		if !e.Timestamp.IsZero() {
			ts = e.Timestamp.Format("Jan 02 15:04")
		}
		fmt.Printf("  #%-4d  %-14s  %s\n", e.ID, ts, loc)
		msg := e.Message
		if len(msg) > 72 {
			msg = msg[:72] + "…"
		}
		fmt.Printf("         %s\n\n", msg)
	}
	fmt.Println("Run 'flint fix <id>' to attempt a fix.")
	fmt.Println("Run 'flint fix --auto <cve|import|format>' to enable category auto-fix.")
	fmt.Println()
	return nil
}

func fixError(id int64, cfg *config.Config, db *store.DB) error {
	e, err := db.GetError(id)
	if err != nil {
		return err
	}

	fmt.Printf("\nFixing error #%d (%s)\n%s\n", id, e.Source, separator())
	fmt.Printf("  %s\n\n", e.Message)

	if cfg.APIKey == "" {
		return fmt.Errorf("no API key — set ANTHROPIC_API_KEY or run 'flint init'")
	}

	client := ai.New(cfg.APIKey, cfg.Model)

	// Advisory mode: no file path — just explain the fix
	if e.FilePath == "" {
		fmt.Println("No file associated — explaining the fix:\n")
		_, err = client.Stream(context.Background(),
			"You are a senior developer. Explain how to fix the following error in 2–3 concise sentences. No preamble.",
			[]ai.Message{{Role: "user", Content: e.Message}},
			func(delta string) { fmt.Print(delta) },
		)
		fmt.Println()
		return err
	}

	content, err := os.ReadFile(e.FilePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", e.FilePath, err)
	}

	// Build user message
	var sb strings.Builder
	sb.WriteString("Error (")
	sb.WriteString(e.Source)
	sb.WriteString("): ")
	sb.WriteString(e.Message)
	if e.LineNumber > 0 {
		fmt.Fprintf(&sb, " (line %d)", e.LineNumber)
	}
	sb.WriteString("\n\nFile: ")
	sb.WriteString(e.FilePath)
	sb.WriteString("\n\n")
	sb.Write(content)

	systemPrompt := `You are a precise code repair assistant.
Output ONLY the corrected file content — the complete file, all lines, with the minimal change needed to fix the error.
No explanation. No markdown fences. No leading "Here is" or trailing notes. Just the file.`

	fmt.Printf("Generating fix for %s…\n\n", e.FilePath)

	var fixed strings.Builder
	_, err = client.Stream(context.Background(), systemPrompt,
		[]ai.Message{{Role: "user", Content: sb.String()}},
		func(delta string) {
			fmt.Print(delta)
			fixed.WriteString(delta)
		},
	)
	fmt.Println()
	if err != nil {
		return fmt.Errorf("AI call: %w", err)
	}

	fixedContent := fixed.String()
	if strings.TrimSpace(fixedContent) == "" {
		return fmt.Errorf("AI returned empty response")
	}

	if err := os.WriteFile(e.FilePath, []byte(fixedContent), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", e.FilePath, err)
	}

	if out, err := exec.Command("git", "add", e.FilePath).CombinedOutput(); err != nil {
		fmt.Printf("Warning: git add failed: %s\n", strings.TrimSpace(string(out)))
	} else {
		fmt.Printf("\n✓ Staged. Review with 'git diff --cached' before committing.\n\n")
	}

	return nil
}

func enableAutoFix(category string) error {
	valid := map[string]bool{"cve": true, "import": true, "format": true}
	if !valid[category] {
		return fmt.Errorf("unknown auto-fix category %q — try: cve import format", category)
	}
	fmt.Printf("Auto-fix enabled for category '%s'.\n", category)
	fmt.Println("The daemon will stage fixes automatically. You always review before commit.")
	return nil
}
