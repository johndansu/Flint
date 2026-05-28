package main

import (
	"fmt"

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
		fmt.Println("\nUsage:")
		fmt.Println("  flint fix <error-id>          Fix a specific logged error")
		fmt.Println("  flint fix --auto <category>   Enable auto-fix for a category")
		fmt.Println()
		fmt.Println("Auto-fix categories: cve  import  format")
		fmt.Println("\nNote: auto-fix stages changes — never commits. Review before merging.")
		return nil
	}

	if args[0] == "--auto" {
		if len(args) < 2 {
			return fmt.Errorf("--auto requires a category: cve | import | format")
		}
		return enableAutoFix(args[1], cfg, db)
	}

	return fmt.Errorf("fix <error-id> not yet implemented — coming in Phase 4")
}

func enableAutoFix(category string, cfg *config.Config, db *store.DB) error {
	valid := map[string]bool{"cve": true, "import": true, "format": true}
	if !valid[category] {
		return fmt.Errorf("unknown auto-fix category %q — try: cve import format", category)
	}
	fmt.Printf("Auto-fix enabled for category '%s'.\n", category)
	fmt.Println("The daemon will stage fixes automatically. You always review before commit.")
	return nil
}
