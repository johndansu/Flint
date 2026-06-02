package main

import (
	"fmt"
	"strings"

	"flint/core/internal/config"
	"flint/core/internal/store"
)

func cmdErrors(args []string) error {
	db, err := store.Open(config.FlintDir())
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer db.Close()

	// Parse flags
	all := false
	filterType := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--all":
			all = true
		case "--type":
			if i+1 < len(args) {
				i++
				filterType = args[i]
			}
		case "--clear":
			return errClear(db)
		}
	}

	limit := 10
	if all {
		limit = 0
	}

	rows, err := db.FilteredErrors(filterType, limit)
	if err != nil {
		return err
	}

	if len(rows) == 0 {
		fmt.Println("No errors logged.")
		fmt.Println("Errors are captured automatically from build failures, test runs, and CVE checks.")
		return nil
	}

	label := "Recent errors"
	if filterType != "" {
		label = fmt.Sprintf("Errors [type: %s]", filterType)
	}
	if all {
		label = "All errors"
	}

	fmt.Printf("\n%s (%d)\n%s\n", label, len(rows), separator())
	for _, e := range rows {
		ts := e.Timestamp.Format("2006-01-02 15:04")
		loc := ""
		if e.FilePath != "" {
			loc = "  " + shortPath(e.FilePath)
			if e.LineNumber > 0 {
				loc += fmt.Sprintf(":%d", e.LineNumber)
			}
		}
		msg := e.Message
		if len(msg) > 90 {
			msg = msg[:90] + "…"
		}
		fmt.Printf("  #%-5d [%s] %-10s %s%s\n", e.ID, ts, e.Source, msg, loc)
	}
	fmt.Println()

	fmt.Println("  Filters: --all  --type <source>  --clear")
	fmt.Println("  Sources: stderr  test  build  runtime  precommit  cve  install")
	fmt.Printf("  Fix:     flint fix <id>\n\n")
	return nil
}

func errClear(db *store.DB) error {
	if err := db.ClearErrors(); err != nil {
		return err
	}
	fmt.Println("Error log cleared.")
	return nil
}

// FilteredErrors is a store helper added here for the command layer.
// The actual implementation lives in store.go.

func errTypeHint() string {
	return strings.Join([]string{"stderr", "test", "build", "runtime", "precommit", "cve", "install"}, "  ")
}
