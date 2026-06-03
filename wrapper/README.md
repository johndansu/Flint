# flint-cli

**The senior dev that never leaves.**

Flint is a passive intelligence layer for your codebase. It watches your work in the background and catches what code review misses — naming drift, security gaps, architectural decay — before you commit.

## Install

```bash
npm install -g flint-cli
```

Or download a binary directly from [Releases](https://github.com/johndansu/Flint/releases).

## Usage

```bash
# Set up Flint in your project
flint init

# Scan staged changes before committing
flint scan

# Check current awareness level and observations
flint status

# Start the background daemon (watches your project continuously)
flint daemon start

# Open the AI REPL with full codebase context
flint repl
```

## How it works

Flint runs a lightweight daemon (`flintd`) that monitors your file edits, git activity, and commit patterns. It builds a model of your codebase over time and surfaces targeted observations — not noise, not alerts, just the things a senior developer would flag.

Three awareness levels control how much it watches:

| Level | Behaviour |
|-------|-----------|
| `spark` | Manual tools only — `flint scan`, `flint repl` |
| `flame` | Daemon watches the current project |
| `forge` | Daemon watches all projects |

```bash
flint config set awareness flame
```

## VS Code Extension

Install the [Flint extension](https://github.com/johndansu/Flint/releases) for inline observations, a sidebar panel, and one-click daemon control.

## Requirements

- Node.js 18+
- macOS, Linux, or Windows (x64 / arm64)
- An Anthropic API key (`ANTHROPIC_API_KEY`)

## License

MIT
