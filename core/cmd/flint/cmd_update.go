package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"flint/core/internal/ai"
	"flint/core/internal/config"
)

func cmdUpdate(args []string) error {
	cfg, err := config.Load(workspaceRoot())
	if err != nil {
		return err
	}
	if cfg.APIKey == "" {
		return fmt.Errorf("no API key — set ANTHROPIC_API_KEY or run 'flint init'")
	}

	format   := "plain"
	audience := "pm"
	n        := 20

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--format":
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		case "--audience":
			if i+1 < len(args) {
				audience = args[i+1]
				i++
			}
		}
	}

	log := recentCommits(n)
	if log == "" {
		return fmt.Errorf("no git commits found")
	}

	system := updateSystem(format, audience)

	fmt.Printf("\nGenerating %s update for %s...\n\n", format, audience)

	client := ai.New(cfg.APIKey, cfg.Model)
	msg := "Recent commits:\n```\n" + log + "\n```"
	_, err = client.Stream(context.Background(), system,
		[]ai.Message{{Role: "user", Content: msg}},
		func(delta string) { fmt.Print(delta) },
	)
	fmt.Println()
	return err
}

func recentCommits(n int) string {
	out, err := exec.Command("git", "log",
		fmt.Sprintf("-%d", n),
		"--pretty=format:%h %s (%an, %ar)",
	).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func updateSystem(format, audience string) string {
	var tone string
	switch audience {
	case "cto":
		tone = "a CTO who cares about architecture, risk, and velocity"
	case "dev":
		tone = "a developer who wants to know what changed and why"
	default:
		tone = "a product manager who needs to understand progress without technical jargon"
	}

	var formatNote string
	switch format {
	case "slack":
		formatNote = "Format as a short Slack message with bullet points. Max 5 bullets. Start with a one-line summary."
	case "email":
		formatNote = "Format as a brief email. Subject line first, then a short body. Professional but conversational."
	default:
		formatNote = "Format as plain text. Short paragraphs. No bullet points needed."
	}

	return fmt.Sprintf(`You are a technical writer summarising recent software development work.
Your audience is %s.
Translate the git commit log into plain English. Focus on what was built and why — not the technical details.
%s
Do not mention commit hashes. Do not use developer jargon unless your audience is a developer.`, tone, formatNote)
}
