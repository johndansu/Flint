package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "flint: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "init":
		return cmdInit(rest)
	case "status":
		return cmdStatus(rest)
	case "scan":
		return cmdScan(rest)
	case "explain":
		return cmdExplain(rest)
	case "diff":
		return cmdDiff(rest)
	case "fix":
		return cmdFix(rest)
	case "watch":
		return cmdWatch(rest)
	case "update":
		return cmdUpdate(rest)
	case "precommit":
		return cmdPreCommit(rest)
	case "version", "--version", "-v":
		fmt.Println("flint v0.1.0")
		return nil
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		// Try to match a role+tool invocation: flint <role> <category>
		return cmdTool(cmd, rest)
	}
}

// sharedDir returns the path to the shared/ directory.
// In production this sits alongside the binary; during dev we walk up to the repo root.
func sharedDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "shared"
	}

	// dev: binary is in core/cmd/flint — walk up three levels to repo root
	dir := filepath.Dir(exe)
	candidate := filepath.Join(dir, "..", "..", "..", "shared")
	if _, err := os.Stat(filepath.Join(candidate, "tools.json")); err == nil {
		return filepath.Clean(candidate)
	}

	// production: shared/ is next to the binary
	candidate = filepath.Join(dir, "shared")
	if _, err := os.Stat(filepath.Join(candidate, "tools.json")); err == nil {
		return candidate
	}

	// fallback: look relative to working directory
	return "shared"
}

func printUsage() {
	fmt.Print(`flint — the senior dev that never leaves

USAGE
  flint init                    Set up Flint for this workspace
  flint status                  Show daemon and awareness status
  flint scan                    Index codebase, establish baselines
  flint explain <file|func>     Explain what something does and why
  flint diff [ref]              Plain-English diff (staged or commit range)
  flint fix [--auto <cat>]      Fix a logged error or enable auto-fix
  flint watch <pattern>         Set a tripwire on a code pattern
  flint update [--format slack] Last N commits as stakeholder update
  flint version                 Print version

TOOL INVOCATION
  flint <role> <category>       Run a manual tool

  Roles:     webdev  fullstack  devops  cyber  data  mobile
  Categories: doc  review  debug  audit

  Examples:
    flint webdev review          Review recent changes as a web developer
    flint devops audit           Audit infrastructure and deployment config
    flint data debug             Debug a data processing problem

  ` + platformNote() + `
`)
}

func platformNote() string {
	if runtime.GOOS == "windows" {
		return "Running on Windows. Daemon features require WSL2 or a Unix socket shim."
	}
	return ""
}
