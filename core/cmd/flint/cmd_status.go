package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"flint/core/internal/config"
	"flint/core/internal/ipc"
	"flint/core/internal/store"
)

func cmdStatus(args []string) error {
	cfg, err := config.Load(workspaceRoot())
	if err != nil {
		return err
	}

	db, err := store.Open(config.FlintDir())
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer db.Close()

	fmt.Printf("\nFlint status\n%s\n", separator())
	fmt.Printf("  Awareness:    %s\n", cfg.Awareness)
	fmt.Printf("  Human layer:  %v\n", cfg.HumanLayer)
	fmt.Printf("  Model:        %s\n", cfg.Model)
	fmt.Printf("  API key:      %s\n", maskKey(cfg.APIKey))
	fmt.Printf("  Flint dir:    %s\n", config.FlintDir())
	fmt.Printf("  Daemon:       %s\n", daemonStatus(cfg))

	if n, err := db.CountObservations(); err == nil {
		fmt.Printf("  Observations: %d total\n", n)
	}

	if errs, err := db.RecentErrors(3); err == nil && len(errs) > 0 {
		fmt.Printf("\n  Recent errors:\n")
		for _, e := range errs {
			loc := e.Source
			if e.FilePath != "" {
				loc += " · " + e.FilePath
				if e.LineNumber > 0 {
					loc += fmt.Sprintf(":%d", e.LineNumber)
				}
			}
			msg := e.Message
			if len(msg) > 60 {
				msg = msg[:60] + "…"
			}
			fmt.Printf("    #%-4d %s\n", e.ID, loc)
			fmt.Printf("          %s\n", msg)
		}
		fmt.Printf("\n  Run 'flint fix <id>' to attempt a fix.\n")
	}

	fmt.Println()
	return nil
}

// daemonStatus tries to connect to the daemon socket and returns a status string.
func daemonStatus(cfg *config.Config) string {
	if cfg.Awareness == config.Spark {
		return "not used (spark mode)"
	}
	socketPath := ipc.SocketPath(config.FlintDir(), workspaceRoot())
	conn, err := net.DialTimeout("unix", socketPath, 200*time.Millisecond)
	if err != nil {
		return "not running  (start with 'flintd')"
	}
	conn.Close()
	return "running"
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
