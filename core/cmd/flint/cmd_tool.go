package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"flint/core/internal/ai"
	"flint/core/internal/config"
	"flint/core/internal/store"
	"flint/core/internal/tools"
)

var validRoles = map[string]bool{
	"webdev": true, "fullstack": true, "devops": true,
	"cyber": true, "data": true, "mobile": true,
}

var validCategories = map[string]bool{
	"doc": true, "review": true, "debug": true, "audit": true,
}

// cmdTool handles: flint <role> <category> [extra context...]
func cmdTool(role string, args []string) error {
	if !validRoles[role] {
		return fmt.Errorf("unknown command or role %q\n\nRun 'flint help' for usage", role)
	}

	if len(args) == 0 {
		listToolsForRole(role)
		return nil
	}

	category := args[0]
	if !validCategories[category] {
		return fmt.Errorf("unknown category %q — try: doc review debug audit", category)
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if cfg.APIKey == "" {
		return fmt.Errorf("no API key found — set ANTHROPIC_API_KEY or run 'flint init'")
	}

	if err := tools.Load(sharedDir()); err != nil {
		return fmt.Errorf("load tools: %w", err)
	}

	toolID := role + "-" + category
	tool, err := tools.ByID(toolID)
	if err != nil {
		return err
	}

	// Gather context: git diff + optional extra args
	userInput, err := gatherInput(args[1:])
	if err != nil {
		return err
	}

	// Record usage for future personalisation
	if db, err := openStore(); err == nil {
		_ = db.RecordToolUsage(toolID, role)
		db.Close()
	}

	fmt.Printf("\n%s — %s\n%s\n\n", tool.Name, tool.Description, strings.Repeat("─", 60))

	client := ai.New(cfg.APIKey, cfg.Model)
	messages := []ai.Message{
		{Role: "user", Content: userInput},
	}

	_, err = client.Stream(context.Background(), tool.Prompt, messages, func(delta string) {
		fmt.Print(delta)
	})
	fmt.Println()
	return err
}

// gatherInput collects context for the tool: staged diff by default, plus any extra text.
func gatherInput(extra []string) (string, error) {
	var parts []string

	// Try to get git diff --staged first, then working tree
	diff := gitDiff("--staged")
	if diff == "" {
		diff = gitDiff("HEAD")
	}
	if diff != "" {
		parts = append(parts, "Git diff:\n```\n"+diff+"\n```")
	}

	if len(extra) > 0 {
		parts = append(parts, strings.Join(extra, " "))
	}

	if len(parts) == 0 {
		// No diff and no args — prompt for input
		fmt.Print("What do you want me to look at? (paste code, describe the problem, or press Enter for current directory): ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				parts = append(parts, line)
			}
		}
	}

	if len(parts) == 0 {
		return "Review the current state of this codebase.", nil
	}

	return strings.Join(parts, "\n\n"), nil
}

func gitDiff(args ...string) string {
	cmdArgs := append([]string{"diff"}, args...)
	out, err := exec.Command("git", cmdArgs...).Output()
	if err != nil {
		return ""
	}
	s := strings.TrimSpace(string(out))
	if len(s) > 40000 {
		// Truncate very large diffs to keep within token budget
		s = s[:40000] + "\n... (truncated)"
	}
	return s
}

func listToolsForRole(role string) {
	if err := tools.Load(sharedDir()); err != nil {
		fmt.Fprintf(os.Stderr, "load tools: %v\n", err)
		return
	}
	ts := tools.ForRole(role)
	fmt.Printf("\nTools for role '%s':\n\n", role)
	for _, t := range ts {
		fmt.Printf("  flint %s %-8s  %s\n", role, t.Category, t.Description)
	}
	fmt.Println()
}

func loadConfig() (*config.Config, error) {
	wd, _ := os.Getwd()
	return config.Load(wd)
}

func openStore() (*store.DB, error) {
	return store.Open(config.FlintDir())
}
