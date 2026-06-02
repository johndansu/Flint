package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"flint/core/internal/config"
	"flint/core/internal/ipc"
	"flint/core/internal/store"
)

func cmdDiagnose(_ []string) error {
	root := workspaceRoot()
	cfg, _ := config.Load(root)

	fmt.Printf("\nFlint diagnose\n%s\n", separator())

	// --- Environment ---
	fmt.Println("  Environment")
	checkItem("OS", runtime.GOOS+"/"+runtime.GOARCH, true)
	checkItem("Flint dir", config.FlintDir(), dirExists(config.FlintDir()))

	// --- Config ---
	fmt.Println("\n  Config")
	checkItem("Awareness", string(cfg.Awareness), true)
	checkItem("Human layer", fmt.Sprintf("%v", cfg.HumanLayer), true)
	checkItem("Model", cfg.Model, true)
	apiOK := cfg.APIKey != ""
	apiVal := "(not set)"
	if apiOK {
		apiVal = cfg.APIKey[:4] + "..." + cfg.APIKey[len(cfg.APIKey)-4:]
	}
	checkItem("API key", apiVal, apiOK)

	// --- Thresholds (only show if any are customised) ---
	t := cfg.Thresholds
	if t.BurnoutHours > 0 || t.SpiralEdits > 0 || t.LoneWolfDays > 0 || t.SessionGapMins > 0 {
		fmt.Println("\n  Thresholds (customised)")
		if t.BurnoutHours > 0 {
			fmt.Printf("    burnoutHours       %d\n", t.BurnoutHours)
		}
		if t.SpiralEdits > 0 {
			fmt.Printf("    spiralEdits        %d\n", t.SpiralEdits)
		}
		if t.SpiralWindowMins > 0 {
			fmt.Printf("    spiralWindowMins   %d\n", t.SpiralWindowMins)
		}
		if t.LoneWolfDays > 0 {
			fmt.Printf("    loneWolfDays       %d\n", t.LoneWolfDays)
		}
		if t.SessionGapMins > 0 {
			fmt.Printf("    sessionGapMins     %d\n", t.SessionGapMins)
		}
		if t.HumanMaxPerSession > 0 {
			fmt.Printf("    humanMaxPerSession %d\n", t.HumanMaxPerSession)
		}
		if t.VibeMaxPerSession > 0 {
			fmt.Printf("    vibeMaxPerSession  %d\n", t.VibeMaxPerSession)
		}
	} else {
		fmt.Println("\n  Thresholds: all defaults")
		fmt.Printf("    burnout=%dh  spiral=%d edits/%dmin  loneWolf=%dd  sessionGap=%dmin\n",
			3, 10, 30, 3, 30)
	}

	// --- Daemon ---
	fmt.Println("\n  Daemon")
	if cfg.Awareness == config.Spark {
		checkItem("Daemon", "not used (spark mode)", true)
	} else {
		sockPath := ipc.SocketPath(config.FlintDir(), root)
		running := socketAlive(sockPath)
		status := "not running"
		if running {
			status = "running"
		}
		checkItem("Daemon", status, running)
		checkItem("Socket", sockPath, true)

		pidFile := daemonPIDPath(root)
		if raw, err := os.ReadFile(pidFile); err == nil {
			checkItem("PID file", strings.TrimSpace(string(raw)), true)
		}
	}

	// --- Git ---
	fmt.Println("\n  Git")
	gitPath, gitErr := exec.LookPath("git")
	checkItem("git binary", gitPath, gitErr == nil)
	if gitErr == nil {
		isRepo := exec.Command("git", "-C", root, "rev-parse", "--git-dir").Run() == nil
		checkItem("workspace is git repo", root, isRepo)
		if isRepo {
			branch := gitOutputDiag(root, "rev-parse", "--abbrev-ref", "HEAD")
			checkItem("current branch", branch, true)
		}
	}

	// --- Dependency files ---
	fmt.Println("\n  Dependency files")
	depFiles := []string{"package.json", "go.mod", "requirements.txt", "Cargo.toml"}
	foundAny := false
	for _, f := range depFiles {
		path := filepath.Join(root, f)
		if _, err := os.Stat(path); err == nil {
			checkItem(f, "found", true)
			foundAny = true
		}
	}
	if !foundAny {
		checkItem("none found", "CVE monitor will skip dep checks", false)
	}

	// --- Store ---
	fmt.Println("\n  Store")
	db, err := store.Open(config.FlintDir())
	if err != nil {
		checkItem("SQLite", err.Error(), false)
	} else {
		defer db.Close()
		n, _ := db.CountObservations()
		checkItem("SQLite", "ok", true)
		checkItem("Observations stored", fmt.Sprintf("%d", n), true)

		if lastCheck, _ := db.GetBaseline("deps.last_check"); lastCheck != "" {
			t, _ := time.Parse(time.RFC3339, lastCheck)
			age := time.Since(t).Round(time.Hour)
			checkItem("Last CVE check", fmt.Sprintf("%v ago", age), true)
		} else {
			checkItem("Last CVE check", "never", false)
		}

		if scannedAt, _ := db.GetBaseline("scan.scanned_at"); scannedAt != "" {
			t, _ := time.Parse(time.RFC3339, scannedAt)
			age := time.Since(t).Round(time.Minute)
			checkItem("Last scan", fmt.Sprintf("%v ago", age), true)
		} else {
			checkItem("Last scan", "never — run 'flint scan'", false)
		}
	}

	fmt.Println()
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func checkItem(label, value string, ok bool) {
	icon := "✓"
	if !ok {
		icon = "✗"
	}
	fmt.Printf("  %s  %-28s %s\n", icon, label, value)
}

func dirExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func socketAlive(path string) bool {
	conn, err := net.DialTimeout("unix", path, 200*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func gitOutputDiag(dir string, args ...string) string {
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
