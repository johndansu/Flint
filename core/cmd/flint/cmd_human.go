package main

import (
	"fmt"
	"strings"

	"flint/core/internal/config"
	"flint/core/internal/store"
)

func cmdHuman(args []string) error {
	sub := ""
	if len(args) > 0 {
		sub = args[0]
	}

	switch sub {
	case "on":
		return humanToggle(true)
	case "off":
		return humanToggle(false)
	case "status":
		return humanStatus()
	case "--clear", "clear":
		return humanClear()
	default:
		fmt.Println("\nUsage:")
		fmt.Println("  flint human on        Enable human intelligence layer")
		fmt.Println("  flint human off       Disable human intelligence layer")
		fmt.Println("  flint human status    Show current state and what it tracks")
		fmt.Println("  flint human --clear   Delete all human-layer observations")
		fmt.Println()
		fmt.Println("  The human layer observes behavioural patterns: burnout signals,")
		fmt.Println("  debugging spirals, and isolation from the rest of the team.")
		fmt.Println("  All data stays local. Nothing is transmitted.")
		return nil
	}
}

func humanToggle(on bool) error {
	root := workspaceRoot()
	cfg, err := config.Load(root)
	if err != nil {
		return err
	}

	cfg.HumanLayer = on
	if err := config.SaveGlobal(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	// Mirror to local .flintrc if it exists
	_ = config.SaveLocal(root, &config.Config{
		Awareness:  cfg.Awareness,
		HumanLayer: on,
	})

	if on {
		fmt.Println("Human intelligence layer enabled.")
		fmt.Println()
		fmt.Println("  Flint will now observe behavioural patterns alongside code quality.")
		fmt.Println("  The daemon must be restarted to pick up this change.")
		fmt.Println("  Run 'flint stop && flint start' to apply.")
	} else {
		fmt.Println("Human intelligence layer disabled.")
		fmt.Println("  Run 'flint stop && flint start' to apply.")
	}
	return nil
}

func humanStatus() error {
	root := workspaceRoot()
	cfg, err := config.Load(root)
	if err != nil {
		return err
	}

	db, err := store.Open(config.FlintDir())
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer db.Close()

	fmt.Printf("\nHuman layer\n%s\n", separator())
	state := "disabled"
	if cfg.HumanLayer {
		state = "enabled"
	}
	fmt.Printf("  State:    %s\n", state)
	fmt.Println()

	if !cfg.HumanLayer {
		fmt.Println("  Enable with: flint human on")
		fmt.Println()
		return nil
	}

	fmt.Println("  Tracking:")
	fmt.Println("    Burnout signal        — coding 3+ hours without a break")
	fmt.Println("    Debugging spiral      — same file edited 10+ times in 30 minutes")
	fmt.Println("    Lone wolf             — no git commits in 3+ days")
	fmt.Println()

	// Show recent human observations
	obs, err := db.HumanObservations(5)
	if err == nil && len(obs) > 0 {
		fmt.Printf("  Recent observations (%d):\n", len(obs))
		for _, o := range obs {
			ts := o.Timestamp.Format("2006-01-02 15:04")
			text := o.Text
			if len(text) > 70 {
				text = text[:70] + "…"
			}
			fmt.Printf("    [%s] %s\n", ts, text)
		}
	} else {
		fmt.Println("  No human observations fired yet.")
		fmt.Println("  The layer needs time to establish patterns — check back after a few sessions.")
	}

	// Show cooldown baselines
	fmt.Println()
	thresholds := cfg.Thresholds
	fmt.Printf("  Thresholds: burnout=%dh  spiral=%d edits/%dmin  loneWolf=%dd\n",
		thresholds.BurnoutHours, thresholds.SpiralEdits, thresholds.SpiralWindowMins, thresholds.LoneWolfDays)
	fmt.Println("  (Override in .flintrc under \"thresholds\")")
	fmt.Println()
	return nil
}

func humanClear() error {
	db, err := store.Open(config.FlintDir())
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer db.Close()

	if err := db.ClearHumanObservations(); err != nil {
		return err
	}
	// Also clear human-layer baselines
	for _, key := range []string{
		"git.stale_branch", "git.fired_msg_hashes",
		"deps.reported_cves", "deps.last_check",
	} {
		_ = db.SetBaseline(key, "")
	}
	fmt.Println("Human intelligence data cleared.")
	fmt.Println("  Observations, cooldown timers, and reported signals reset.")
	return nil
}

var _ = strings.TrimSpace // avoid unused import
