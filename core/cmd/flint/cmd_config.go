package main

import (
	"fmt"
	"strconv"
	"strings"

	"flint/core/internal/config"
)

func cmdConfig(args []string) error {
	if len(args) == 0 {
		return configShow()
	}

	// flint config <key> <value>
	if len(args) == 2 {
		return configSet(args[0], args[1])
	}

	fmt.Println("\nUsage:")
	fmt.Println("  flint config                 Show current configuration")
	fmt.Println("  flint config <key> <value>   Set a configuration value")
	fmt.Println()
	fmt.Println("  Keys:")
	fmt.Println("    awareness          spark | flame | forge")
	fmt.Println("    model              Claude model ID")
	fmt.Println("    human_layer        true | false")
	fmt.Println("    burnout_hours      integer (default 3)")
	fmt.Println("    spiral_edits       integer (default 10)")
	fmt.Println("    spiral_window_mins integer (default 30)")
	fmt.Println("    lone_wolf_days     integer (default 3)")
	fmt.Println()
	return nil
}

func cmdAwareness(args []string) error {
	if len(args) == 0 {
		root := workspaceRoot()
		cfg, err := config.Load(root)
		if err != nil {
			return err
		}
		fmt.Printf("Awareness: %s\n", cfg.Awareness)
		fmt.Println("  Levels: spark  flame  forge")
		fmt.Println("  Set with: flint awareness <level>")
		return nil
	}

	level := strings.ToLower(args[0])
	switch level {
	case "spark", "flame", "forge":
	default:
		return fmt.Errorf("unknown awareness level %q — use: spark, flame, forge", args[0])
	}

	root := workspaceRoot()
	cfg, err := config.Load(root)
	if err != nil {
		return err
	}

	cfg.Awareness = config.AwarenessLevel(level)
	if err := config.SaveGlobal(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	_ = config.SaveLocal(root, &config.Config{Awareness: config.AwarenessLevel(level), HumanLayer: cfg.HumanLayer})

	fmt.Printf("Awareness set to %s.\n", level)
	switch level {
	case "spark":
		fmt.Println("  Manual tools only. Daemon is not involved.")
	case "flame":
		fmt.Println("  Daemon watches the current workspace.")
	case "forge":
		fmt.Println("  Daemon watches all registered workspaces.")
	}
	fmt.Println("  Run 'flint stop && flint start' to apply if daemon is running.")
	return nil
}

func configShow() error {
	root := workspaceRoot()
	cfg, err := config.Load(root)
	if err != nil {
		return err
	}

	humanState := "disabled"
	if cfg.HumanLayer {
		humanState = "enabled"
	}
	model := cfg.Model
	if model == "" {
		model = "(default)"
	}
	apiKeyState := "(not set)"
	if cfg.APIKey != "" {
		apiKeyState = "(set)"
	}

	fmt.Printf("\nFlint config\n%s\n", separator())
	fmt.Printf("  awareness          %s\n", cfg.Awareness)
	fmt.Printf("  model              %s\n", model)
	fmt.Printf("  api_key            %s\n", apiKeyState)
	fmt.Printf("  human_layer        %s\n", humanState)
	fmt.Println()
	fmt.Printf("  Thresholds (0 = default):\n")
	t := cfg.Thresholds
	fmt.Printf("    burnout_hours          %d\n", t.BurnoutHours)
	fmt.Printf("    spiral_edits           %d\n", t.SpiralEdits)
	fmt.Printf("    spiral_window_mins     %d\n", t.SpiralWindowMins)
	fmt.Printf("    lone_wolf_days         %d\n", t.LoneWolfDays)
	fmt.Println()
	fmt.Println("  Edit .flintrc in your workspace to override thresholds.")
	fmt.Println("  Use 'flint config <key> <value>' to set values.")
	fmt.Println()
	return nil
}

func configSet(key, value string) error {
	root := workspaceRoot()
	cfg, err := config.Load(root)
	if err != nil {
		return err
	}

	switch key {
	case "awareness":
		return cmdAwareness([]string{value})
	case "human_layer":
		on := value == "true" || value == "1" || value == "yes"
		cfg.HumanLayer = on
	case "model":
		cfg.Model = value
	case "burnout_hours":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("burnout_hours must be an integer")
		}
		cfg.Thresholds.BurnoutHours = n
	case "spiral_edits":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("spiral_edits must be an integer")
		}
		cfg.Thresholds.SpiralEdits = n
	case "spiral_window_mins":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("spiral_window_mins must be an integer")
		}
		cfg.Thresholds.SpiralWindowMins = n
	case "lone_wolf_days":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("lone_wolf_days must be an integer")
		}
		cfg.Thresholds.LoneWolfDays = n
	default:
		return fmt.Errorf("unknown config key %q — run 'flint config' for valid keys", key)
	}

	if err := config.SaveGlobal(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	_ = config.SaveLocal(root, cfg)
	fmt.Printf("Set %s = %s\n", key, value)
	return nil
}
