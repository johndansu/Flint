package main

import (
	"fmt"
	"time"

	"flint/core/internal/config"
	"flint/core/internal/store"

	"github.com/google/uuid"
)

func cmdWatch(args []string) error {
	if len(args) == 0 {
		fmt.Println("\nUsage:")
		fmt.Println("  flint watch <pattern>    Set a tripwire (e.g. flint watch \"payment processing\")")
		fmt.Println("  flint watch list         List active tripwires")
		fmt.Println("  flint watch remove <id>  Remove a tripwire")
		return nil
	}

	cfg, err := config.Load(workspaceRoot())
	if err != nil {
		return err
	}
	_ = cfg

	db, err := store.Open(config.FlintDir())
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer db.Close()

	switch args[0] {
	case "list":
		return watchList(db)
	case "remove", "rm":
		if len(args) < 2 {
			return fmt.Errorf("watch remove requires an id")
		}
		return watchRemove(db, args[1])
	default:
		return watchAdd(db, args[0])
	}
}

func watchAdd(db *store.DB, pattern string) error {
	tw := store.Tripwire{
		ID:        uuid.New().String(),
		Pattern:   pattern,
		CreatedAt: time.Now(),
		Active:    true,
	}
	if err := db.InsertTripwire(tw); err != nil {
		return fmt.Errorf("add tripwire: %w", err)
	}
	fmt.Printf("✓ Tripwire set: %q\n", pattern)
	fmt.Printf("  ID: %s\n", tw.ID)
	fmt.Println("  The daemon will fire immediately when any file touching this pattern changes.")
	return nil
}

func watchList(db *store.DB) error {
	tws, err := db.ActiveTripwires()
	if err != nil {
		return err
	}
	if len(tws) == 0 {
		fmt.Println("No active tripwires. Set one with: flint watch <pattern>")
		return nil
	}
	fmt.Printf("\nActive tripwires (%d):\n\n", len(tws))
	for _, tw := range tws {
		fmt.Printf("  %s  %q\n", tw.ID[:8], tw.Pattern)
	}
	fmt.Println()
	return nil
}

func watchRemove(db *store.DB, id string) error {
	if err := db.DeactivateTripwire(id); err != nil {
		return err
	}
	fmt.Printf("✓ Tripwire %s removed.\n", id)
	return nil
}
