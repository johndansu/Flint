package main

import (
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"flint/core/internal/config"
	"flint/core/internal/store"
)



func cmdScan(args []string) error {
	cfg, err := config.Load(workspaceRoot())
	if err != nil {
		return err
	}

	db, err := store.Open(config.FlintDir())
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer db.Close()

	root := workspaceRoot()
	fmt.Printf("\nScanning %s...\n", root)

	stats := scanWorkspace(root)

	// Persist baselines
	now := time.Now().UTC().Format(time.RFC3339)
	baselines := map[string]string{
		"scan.root":        root,
		"scan.files":       strconv.Itoa(stats.files),
		"scan.go_files":    strconv.Itoa(stats.goFiles),
		"scan.ts_files":    strconv.Itoa(stats.tsFiles),
		"scan.python_files": strconv.Itoa(stats.pyFiles),
		"scan.scanned_at":  now,
	}

	if project := cfg.Project; project == "" {
		// Default project name to directory basename
		baselines["project.name"] = filepath.Base(root)
	}

	// Try to get git remote for project identification
	if remote := gitRemote(); remote != "" {
		baselines["git.remote"] = remote
	}

	raw := db.Raw()
	for k, v := range baselines {
		_, err := raw.Exec(
			`INSERT INTO baselines (key, value, updated_at) VALUES (?,?,?)
			 ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at`,
			k, v, now,
		)
		if err != nil {
			return fmt.Errorf("save baseline %s: %w", k, err)
		}
	}

	fmt.Printf("✓ Indexed %d files (%d Go, %d TypeScript, %d Python)\n",
		stats.files, stats.goFiles, stats.tsFiles, stats.pyFiles)
	fmt.Println("  Baselines established. Flint now knows what's normal for this project.")
	fmt.Println()

	return nil
}

type scanStats struct {
	files, goFiles, tsFiles, pyFiles int
}

var skipDirs = map[string]bool{
	"node_modules": true, ".git": true, "vendor": true,
	".cache": true, "dist": true, "build": true, "out": true,
}

func scanWorkspace(root string) scanStats {
	var stats scanStats
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && skipDirs[d.Name()] {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		stats.files++
		switch strings.ToLower(filepath.Ext(d.Name())) {
		case ".go":
			stats.goFiles++
		case ".ts", ".tsx":
			stats.tsFiles++
		case ".py":
			stats.pyFiles++
		}
		return nil
	})
	return stats
}

func gitRemote() string {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

