package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"flint/core/internal/ai"
	"flint/core/internal/config"
)

const diffSystem = `You are a senior software engineer reviewing a git diff.
Explain what changed, the apparent intent behind the changes, and any risks introduced.
Be specific — name the functions and files affected. Flag anything that looks incomplete or risky.
End with one sentence on the biggest risk in this diff, or "No significant risks" if there are none.`

func cmdDiff(args []string) error {
	cfg, err := config.Load(workspaceRoot())
	if err != nil {
		return err
	}
	if cfg.APIKey == "" {
		return fmt.Errorf("no API key — set ANTHROPIC_API_KEY or run 'flint init'")
	}

	diff := resolveDiff(args)
	if diff == "" {
		return fmt.Errorf("no diff found — stage some changes or provide a ref (e.g. flint diff HEAD~1)")
	}

	if len(diff) > 40000 {
		diff = diff[:40000] + "\n... (truncated)"
	}

	fmt.Println("\nReading diff...")

	client := ai.New(cfg.APIKey, cfg.Model)
	msg := "Git diff:\n```\n" + diff + "\n```"
	_, err = client.Stream(context.Background(), diffSystem,
		[]ai.Message{{Role: "user", Content: msg}},
		func(delta string) { fmt.Print(delta) },
	)
	fmt.Println()
	return err
}

func resolveDiff(args []string) string {
	if len(args) > 0 {
		// User supplied a ref range like HEAD~1 or main..feature
		ref := args[0]
		out, err := exec.Command("git", "diff", ref).Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	}

	// Try staged first, fall back to working tree vs HEAD
	if d := gitDiff("--staged"); d != "" {
		return d
	}
	return gitDiff("HEAD")
}
