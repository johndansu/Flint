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

	// Resolve shared/ directory relative to the binary (same layout as flint CLI)
	exe, _ := os.Executable()
	sharedDir := filepath.Join(filepath.Dir(exe), "..", "..", "..", "shared")
	if _, err := os.Stat(filepath.Join(sharedDir, "footguns.json")); err != nil {
		sharedDir = "shared" // fallback to working directory
	}

	socketPath := ipc.SocketPath(config.FlintDir(), workspaceRoot)
	d, err := daemon.New(cfg, db, socketPath, sharedDir)
	if err != nil {
		return fmt.Errorf("daemon: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown on SIGINT / SIGTERM
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Println("shutting down...")
		cancel()
	}()

	log.Printf("socket: %s", socketPath)
	return d.Run(ctx, workspaceRoot)
}
