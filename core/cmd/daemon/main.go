package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"flint/core/internal/config"
	"flint/core/internal/daemon"
	"flint/core/internal/ipc"
	"flint/core/internal/registry"
	"flint/core/internal/store"
)

func main() {
	log.SetPrefix("flintd: ")
	log.SetFlags(log.Ltime)

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "flintd: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	workspaceRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %w", err)
	}

	cfg, err := config.Load(workspaceRoot)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	if cfg.Awareness == config.Spark {
		return fmt.Errorf("awareness is set to 'spark' — daemon not needed. Run 'flint init' to upgrade.")
	}

	db, err := store.Open(config.FlintDir())
	if err != nil {
		return fmt.Errorf("store: %w", err)
	}
	defer db.Close()

	// Resolve shared/ directory relative to the binary
	exe, _ := os.Executable()
	sharedDir := filepath.Join(filepath.Dir(exe), "..", "..", "..", "shared")
	if _, err := os.Stat(filepath.Join(sharedDir, "footguns.json")); err != nil {
		sharedDir = "shared"
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Println("shutting down...")
		cancel()
	}()

	// Forge mode: watch all registered workspaces
	if cfg.Awareness == config.Forge {
		return runForge(ctx, cfg, db, sharedDir, workspaceRoot)
	}

	// Flame mode: watch only the current workspace
	socketPath := ipc.SocketPath(config.FlintDir(), workspaceRoot)
	d, err := daemon.New(cfg, db, socketPath, sharedDir, workspaceRoot)
	if err != nil {
		return fmt.Errorf("daemon: %w", err)
	}

	log.Printf("socket: %s", socketPath)
	return d.Run(ctx, workspaceRoot)
}

func runForge(ctx context.Context, cfg *config.Config, db *store.DB, sharedDir, currentRoot string) error {
	// Always ensure the current workspace is registered
	_ = registry.Register(config.FlintDir(), currentRoot)

	reg, err := registry.Load(config.FlintDir())
	if err != nil || len(reg.Workspaces) == 0 {
		reg = &registry.Registry{Workspaces: []string{currentRoot}}
	}

	log.Printf("forge mode: %d registered workspace(s)", len(reg.Workspaces))
	for _, w := range reg.Workspaces {
		log.Printf("  %s", w)
	}

	fd := daemon.NewForge(cfg, db, sharedDir)
	return fd.Run(ctx, reg.Workspaces)
}
