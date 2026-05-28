package daemon

import (
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

// agentProcessNames are lowercase substrings of known AI coding tool executables.
var agentProcessNames = []string{
	"claude",   // Claude Code CLI
	"cursor",   // Cursor editor
	"windsurf", // Windsurf editor
	"aider",    // Aider AI pair programmer
	"continue", // Continue VS Code extension helper
}

// processListCache caches the result of the expensive OS process list call.
var (
	procCacheMu  sync.Mutex
	procCache    bool
	procCachedAt time.Time
)

const procCacheTTL = 30 * time.Second

// detectAgentToolActive checks whether a known AI coding tool is currently
// running. The result is cached for 30 seconds to avoid hammering the OS.
func detectAgentToolActive() bool {
	procCacheMu.Lock()
	defer procCacheMu.Unlock()

	if time.Since(procCachedAt) < procCacheTTL {
		return procCache
	}

	procCache = checkAgentProcesses()
	procCachedAt = time.Now()
	return procCache
}

func checkAgentProcesses() bool {
	var out []byte
	var err error

	switch runtime.GOOS {
	case "windows":
		// tasklist /FO CSV /NH: "process.exe","pid","session","num","mem"
		out, err = exec.Command("tasklist", "/FO", "CSV", "/NH").Output()
	default:
		// ps -A -o comm= lists just the command name for every process
		out, err = exec.Command("ps", "-A", "-o", "comm=").Output()
	}
	if err != nil {
		return false
	}

	lower := strings.ToLower(string(out))
	for _, name := range agentProcessNames {
		if strings.Contains(lower, name) {
			return true
		}
	}
	return false
}

// detectMultipleAuthors returns true if more than one git author has committed
// to workspaceRoot within the last 5 minutes — a strong signal for pair
// programming or mob programming.
func detectMultipleAuthors(workspaceRoot string) bool {
	since := time.Now().Add(-5 * time.Minute).Format("2006-01-02T15:04:05")
	out, err := exec.Command(
		"git", "-C", workspaceRoot,
		"log", "--format=%ae", "--since="+since,
	).Output()
	if err != nil {
		return false // not a git repo or git not installed
	}

	seen := make(map[string]struct{})
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if email := strings.TrimSpace(line); email != "" {
			seen[email] = struct{}{}
		}
	}
	return len(seen) >= 2
}
