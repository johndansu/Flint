package main

import (
	"fmt"
	"strings"

	"flint/core/internal/config"
	"flint/core/internal/store"
)

func cmdDismiss(args []string) error {
	if len(args) == 0 {
		fmt.Println("\nUsage:")
		fmt.Println("  flint dismiss <id>   Dismiss an observation (use 'flint memory obs' to see IDs)")
		fmt.Println()
		fmt.Println("  If the observation maps to a tunable threshold, Flint will offer to")
		fmt.Println("  raise the threshold so the same signal fires less often.")
		fmt.Println()
		return nil
	}

	prefix := strings.TrimSpace(args[0])

	db, err := store.Open(config.FlintDir())
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer db.Close()

	obs, err := db.GetObservationByPrefix(prefix)
	if err != nil {
		return err
	}

	fmt.Printf("\nObservation %s\n%s\n", obs.ID[:min8(len(obs.ID))], separator())
	fmt.Printf("  Kind:    %s/%s\n", obs.Kind, obs.Category)
	fmt.Printf("  When:    %s\n", obs.Timestamp.Format("2006-01-02 15:04"))
	text := obs.Text
	if len(text) > 80 {
		text = text[:80] + "…"
	}
	fmt.Printf("  Text:    %s\n", text)
	fmt.Println()

	// Determine if there's a threshold to offer
	suggestion := calibrationSuggestion(obs.Kind, obs.Category)

	if suggestion.key == "" {
		// Just dismiss — no threshold maps to this kind
		if err := db.InsertDismissal(obs.ID, "dismissed"); err != nil {
			return err
		}
		fmt.Println("Dismissed.")
		return nil
	}

	// Offer calibration
	fmt.Printf("  This signal maps to threshold: %s (current: %s)\n", suggestion.label, suggestion.currentStr)
	fmt.Printf("  Dismiss only, or raise the threshold so this fires less?\n")
	fmt.Printf("  [1] Dismiss only\n")
	fmt.Printf("  [2] Dismiss + raise %s to %s\n", suggestion.label, suggestion.proposedStr)
	fmt.Printf("  [3] Cancel\n")
	fmt.Print("> ")

	answer := strings.TrimSpace(readLine())
	switch answer {
	case "1":
		if err := db.InsertDismissal(obs.ID, "dismissed"); err != nil {
			return err
		}
		fmt.Println("Dismissed.")

	case "2":
		if err := db.InsertDismissal(obs.ID, "calibrated"); err != nil {
			return err
		}
		if err := applyCalibration(suggestion); err != nil {
			return err
		}
		fmt.Printf("Dismissed and threshold raised: %s → %s\n", suggestion.currentStr, suggestion.proposedStr)
		fmt.Println("Run 'flint stop && flint start' to apply to the running daemon.")

	default:
		fmt.Println("Cancelled.")
	}

	return nil
}

// ---------------------------------------------------------------------------
// Calibration logic
// ---------------------------------------------------------------------------

type calibration struct {
	key         string // config field name
	label       string // human-readable threshold name
	currentVal  int
	proposedVal int
	currentStr  string
	proposedStr string
}

func calibrationSuggestion(kind, category string) calibration {
	if kind != "human" {
		return calibration{}
	}

	root := workspaceRoot()
	cfg, err := config.Load(root)
	if err != nil {
		return calibration{}
	}
	t := cfg.Thresholds

	switch category {
	case "burnout":
		cur := t.BurnoutHours
		if cur <= 0 {
			cur = 3 // default
		}
		proposed := cur + 1
		return calibration{
			key:         "burnout_hours",
			label:       "burnout_hours",
			currentVal:  cur,
			proposedVal: proposed,
			currentStr:  fmt.Sprintf("%dh", cur),
			proposedStr: fmt.Sprintf("%dh", proposed),
		}

	case "spiral":
		cur := t.SpiralEdits
		if cur <= 0 {
			cur = 10 // default
		}
		proposed := cur + 3
		return calibration{
			key:         "spiral_edits",
			label:       "spiral_edits",
			currentVal:  cur,
			proposedVal: proposed,
			currentStr:  fmt.Sprintf("%d edits", cur),
			proposedStr: fmt.Sprintf("%d edits", proposed),
		}

	case "lone_wolf":
		cur := t.LoneWolfDays
		if cur <= 0 {
			cur = 3 // default
		}
		proposed := cur + 1
		return calibration{
			key:         "lone_wolf_days",
			label:       "lone_wolf_days",
			currentVal:  cur,
			proposedVal: proposed,
			currentStr:  fmt.Sprintf("%dd", cur),
			proposedStr: fmt.Sprintf("%dd", proposed),
		}
	}

	return calibration{}
}

func applyCalibration(c calibration) error {
	root := workspaceRoot()
	cfg, err := config.Load(root)
	if err != nil {
		return err
	}

	switch c.key {
	case "burnout_hours":
		cfg.Thresholds.BurnoutHours = c.proposedVal
	case "spiral_edits":
		cfg.Thresholds.SpiralEdits = c.proposedVal
	case "lone_wolf_days":
		cfg.Thresholds.LoneWolfDays = c.proposedVal
	}

	if err := config.SaveGlobal(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	_ = config.SaveLocal(root, cfg)
	return nil
}

func min8(n int) int {
	if n < 8 {
		return n
	}
	return 8
}
