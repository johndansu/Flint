package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"flint/core/internal/config"
)

func cmdInit(args []string) error {
	fmt.Println("\nWelcome to Flint — the senior dev that never leaves.")

	cfg, _ := config.Load("")

	cfg.Awareness = promptAwareness()
	cfg.APIKey    = promptAPIKey(cfg.APIKey)

	if err := config.SaveGlobal(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	// Write a minimal .flintrc in the current workspace
	wd, _ := os.Getwd()
	local := &config.Config{Awareness: cfg.Awareness}
	if err := config.SaveLocal(wd, local); err != nil {
		return fmt.Errorf("save .flintrc: %w", err)
	}

	fmt.Printf("\n✓ Flint configured (awareness: %s)\n", cfg.Awareness)
	fmt.Printf("  Global config: %s/config.json\n", config.FlintDir())
	fmt.Printf("  Workspace:     %s/.flintrc\n\n", wd)

	switch cfg.Awareness {
	case config.Flame, config.Forge:
		fmt.Println("Run 'flint scan' to index your codebase and establish baselines.")
		fmt.Println("Run 'flintd' to start the background daemon.")
		offerPreCommitHook(wd)
	default:
		fmt.Println("Spark mode: manual tools are ready. Run 'flint help' to see them.")
	}
	fmt.Println()

	return nil
}

// offerPreCommitHook installs the git pre-commit hook if we're in a git repo
// and the hook doesn't already exist.
func offerPreCommitHook(workspaceRoot string) {
	gitDir := filepath.Join(workspaceRoot, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return // not a git repo
	}

	hookPath := filepath.Join(gitDir, "hooks", "pre-commit")
	if _, err := os.Stat(hookPath); err == nil {
		// Hook already exists — don't overwrite
		fmt.Printf("\n  Pre-commit hook already exists at %s\n", hookPath)
		fmt.Println("  Add 'flint precommit' to it manually if it isn't already there.")
		return
	}

	fmt.Print("\nInstall Flint pre-commit checkpoint? [Y/n]: ")
	answer := strings.TrimSpace(strings.ToLower(readLine()))
	if answer == "n" || answer == "no" {
		return
	}

	script := "#!/bin/sh\nflint precommit\n"
	if err := os.MkdirAll(filepath.Dir(hookPath), 0o755); err != nil {
		fmt.Printf("  Could not create hooks dir: %v\n", err)
		return
	}
	if err := os.WriteFile(hookPath, []byte(script), 0o755); err != nil {
		fmt.Printf("  Could not install hook: %v\n", err)
		return
	}
	fmt.Printf("  ✓ Pre-commit hook installed at %s\n", hookPath)
}

func promptAwareness() config.AwarenessLevel {
	fmt.Println("Awareness level:")
	fmt.Println("  1) spark  — manual tools only, no background process")
	fmt.Println("  2) flame  — daemon watches this project (recommended)")
	fmt.Println("  3) forge  — daemon watches everything")
	fmt.Print("Choose [1/2/3] (default: 1 spark): ")

	line := readLine()
	switch strings.TrimSpace(line) {
	case "2", "flame":
		return config.Flame
	case "3", "forge":
		return config.Forge
	default:
		return config.Spark
	}
}

func promptAPIKey(existing string) string {
	if existing != "" {
		fmt.Printf("API key: (already set — press Enter to keep) ")
		line := readLine()
		if strings.TrimSpace(line) == "" {
			return existing
		}
		return strings.TrimSpace(line)
	}

	fmt.Print("Anthropic API key (or set ANTHROPIC_API_KEY): ")
	line := strings.TrimSpace(readLine())
	return line
}

func readLine() string {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}
	return ""
}
