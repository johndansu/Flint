package main

import (
	"fmt"
	"os"

	"flint/core/internal/config"
)

func cmdStatus(args []string) error {
	cfg, err := config.Load(workspaceRoot())
	if err != nil {
		return err
	}

	fmt.Printf("\nFlint status\n%s\n", separator())
	fmt.Printf("  Awareness:    %s\n", cfg.Awareness)
	fmt.Printf("  Human layer:  %v\n", cfg.HumanLayer)
	fmt.Printf("  Model:        %s\n", cfg.Model)
	fmt.Printf("  API key:      %s\n", maskKey(cfg.APIKey))
	fmt.Printf("  Flint dir:    %s\n", config.FlintDir())

	// Daemon connection check — Phase 4 will wire this properly
	fmt.Printf("  Daemon:       not running (start with 'flintd')\n")
	fmt.Println()

	return nil
}

func workspaceRoot() string {
	wd, _ := os.Getwd()
	return wd
}

func maskKey(key string) string {
	if len(key) < 8 {
		return "(not set)"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

func separator() string {
	return "─────────────────────────────────"
}
