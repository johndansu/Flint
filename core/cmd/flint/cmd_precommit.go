package main

import (
	"context"
	"fmt"
	"os"

	"flint/core/internal/ai"
	"flint/core/internal/config"
	"flint/core/internal/errorlog"
	"flint/core/internal/store"
)

const preCommitSystem = `You are a senior software engineer doing a pre-commit consequence analysis.
Given a git diff, identify:
1. Downstream impact — what else could break because of this change
2. Untested paths — changes that have no corresponding test coverage
3. Injection risks — any user input reaching dangerous functions without validation
4. Anything that looks incomplete or half-done

Be specific. Name files and functions. Never block the commit — this is advisory only.
Format: three sections with short bullet points. If a section has nothing to report, write "None."
End with: RISK LEVEL: low | medium | high — and one sentence explaining why.`

// cmdPreCommit is invoked by the git pre-commit hook (never directly by the user).
// It writes to stderr so git captures and displays the output, and never exits non-zero
// (advisory only — never blocks).
func cmdPreCommit(args []string) error {
	cfg, err := config.Load(workspaceRoot())
	if err != nil {
		// Don't block commits due to config errors
		return nil
	}

	if cfg.APIKey == "" {
		return nil // no key, skip silently
	}

	diff := gitDiff("--staged")
	if diff == "" {
		return nil // nothing staged
	}

	if len(diff) > 40000 {
		diff = diff[:40000] + "\n... (truncated)"
	}

	db, _ := store.Open(config.FlintDir())
	if db != nil {
		defer db.Close()
	}

	fmt.Fprintln(os.Stderr, "\nFlint pre-commit checkpoint...")

	client := ai.New(cfg.APIKey, cfg.Model)
	msg := "Staged diff:\n```\n" + diff + "\n```"

	result, err := client.Complete(context.Background(), preCommitSystem,
		[]ai.Message{{Role: "user", Content: msg}},
	)
	if err != nil {
		// Log the failure but don't block the commit
		if db != nil {
			el, lerr := errorlog.New(db, config.FlintDir(), workspaceRoot())
			if lerr == nil {
				defer el.Close()
				_ = el.PreCommit("pre-commit API call failed: " + err.Error())
			}
		}
		return nil
	}

	fmt.Fprintln(os.Stderr, "\n"+result+"\n")
	return nil // never exit non-zero — advisory only
}

// installPreCommitHook writes the git pre-commit hook script to .git/hooks/pre-commit.
func installPreCommitHook(gitRoot string) error {
	hookPath := gitRoot + "/.git/hooks/pre-commit"

	// Don't overwrite an existing hook
	if _, err := os.Stat(hookPath); err == nil {
		fmt.Printf("Pre-commit hook already exists at %s\n", hookPath)
		fmt.Println("Add 'flint precommit' to it manually.")
		return nil
	}

	script := "#!/bin/sh\nflint precommit\n"
	if err := os.WriteFile(hookPath, []byte(script), 0o755); err != nil {
		return fmt.Errorf("install hook: %w", err)
	}
	fmt.Printf("✓ Pre-commit hook installed at %s\n", hookPath)
	return nil
}
