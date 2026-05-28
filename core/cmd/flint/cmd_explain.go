package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"flint/core/internal/ai"
	"flint/core/internal/config"
)

const explainSystem = `You are a senior software engineer reading unfamiliar code.
Explain what the code does, why it exists, what assumptions it makes, and what would break if it were changed.
Be concrete and specific. Comprehension assistance, not code review.
End with one sentence on the most important thing to know before touching this code.`

func cmdExplain(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: flint explain <file> [function]")
	}

	target := args[0]
	content, err := readTarget(target)
	if err != nil {
		return err
	}

	cfg, err := config.Load(workspaceRoot())
	if err != nil {
		return err
	}
	if cfg.APIKey == "" {
		return fmt.Errorf("no API key — set ANTHROPIC_API_KEY or run 'flint init'")
	}

	var userMsg strings.Builder
	userMsg.WriteString(fmt.Sprintf("File: %s\n\n", target))
	if len(args) > 1 {
		userMsg.WriteString(fmt.Sprintf("Focus on: %s\n\n", strings.Join(args[1:], " ")))
	}
	userMsg.WriteString("```\n")
	userMsg.WriteString(content)
	userMsg.WriteString("\n```")

	fmt.Printf("\nExplaining %s...\n\n", target)

	client := ai.New(cfg.APIKey, cfg.Model)
	_, err = client.Stream(context.Background(), explainSystem,
		[]ai.Message{{Role: "user", Content: userMsg.String()}},
		func(delta string) { fmt.Print(delta) },
	)
	fmt.Println()
	return err
}

func readTarget(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("explain: cannot read %s: %w", path, err)
	}
	content := string(b)
	// Cap at ~50K chars to stay well within token limits
	if len(content) > 50000 {
		content = content[:50000] + "\n... (file truncated)"
	}
	return content, nil
}
