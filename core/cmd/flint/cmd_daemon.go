package main

import (
	"crypto/md5"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"flint/core/internal/config"
	"flint/core/internal/ipc"
)

func cmdStart(_ []string) error {
	cfg, err := config.Load(workspaceRoot())
	if err != nil {
		return err
	}
	if cfg.Awareness == config.Spark {
		return fmt.Errorf("awareness is 'spark' — daemon not used. Run 'flint init' to change awareness level")
	}

	root := workspaceRoot()
	sockPath := ipc.SocketPath(config.FlintDir(), root)

	// Already running?
	if conn, err := net.DialTimeout("unix", sockPath, 200*time.Millisecond); err == nil {
		conn.Close()
		fmt.Println("Flint daemon is already running.")
		return nil
	}

	// Find flintd binary
	flintd, err := findFlintd()
	if err != nil {
		return fmt.Errorf("cannot find flintd: %w\n  Build it with: go build -o flintd ./core/cmd/daemon", err)
	}

	pidFile := daemonPIDPath(root)

	// Launch detached
	cmd := exec.Command(flintd)
	cmd.Dir = root
	cmd.Stdout = nil
	cmd.Stderr = nil
	if runtime.GOOS == "windows" {
		// On Windows, CREATE_NEW_PROCESS_GROUP detaches the child from the parent console
		detachProcess(cmd)
	} else {
		detachProcess(cmd)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start daemon: %w", err)
	}

	// Write PID file
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0644); err != nil {
		// Non-fatal — daemon is still running
		fmt.Printf("warning: could not write PID file: %v\n", err)
	}

	// Wait briefly to confirm it stayed alive
	time.Sleep(300 * time.Millisecond)
	if conn, err := net.DialTimeout("unix", sockPath, 500*time.Millisecond); err == nil {
		conn.Close()
		fmt.Printf("Flint daemon started (PID %d).\n", cmd.Process.Pid)
	} else {
		fmt.Printf("Flint daemon launched (PID %d) — socket not yet ready, it may take a moment.\n", cmd.Process.Pid)
	}

	return nil
}

func cmdStop(_ []string) error {
	root := workspaceRoot()
	pidFile := daemonPIDPath(root)

	raw, err := os.ReadFile(pidFile)
	if err != nil {
		// Try to detect via socket even without PID file
		sockPath := ipc.SocketPath(config.FlintDir(), root)
		if _, err2 := net.DialTimeout("unix", sockPath, 200*time.Millisecond); err2 != nil {
			fmt.Println("Flint daemon is not running.")
			return nil
		}
		return fmt.Errorf("daemon appears to be running but no PID file found at %s", pidFile)
	}

	pidStr := strings.TrimSpace(string(raw))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return fmt.Errorf("invalid PID in file: %q", pidStr)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("Process %d not found — may have already exited.\n", pid)
		_ = os.Remove(pidFile)
		return nil
	}

	if err := killProcess(proc); err != nil {
		return fmt.Errorf("stop daemon (PID %d): %w", pid, err)
	}

	_ = os.Remove(pidFile)
	fmt.Printf("Flint daemon stopped (PID %d).\n", pid)
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// findFlintd looks for the flintd binary next to the running flint binary,
// then falls back to PATH.
func findFlintd() (string, error) {
	name := "flintd"
	if runtime.GOOS == "windows" {
		name = "flintd.exe"
	}

	// Prefer sibling binary
	self, err := os.Executable()
	if err == nil {
		candidate := filepath.Join(filepath.Dir(self), name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	// Fall back to PATH
	p, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("%s not found next to flint or in PATH", name)
	}
	return p, nil
}

// daemonPIDPath returns ~/.flint/daemon-<hash>.pid for the given workspace root.
func daemonPIDPath(root string) string {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(root)))[:8]
	return filepath.Join(config.FlintDir(), fmt.Sprintf("daemon-%s.pid", hash))
}
