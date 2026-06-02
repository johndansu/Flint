# Flint

**The senior dev that never leaves.**

Flint watches your code, git activity, and dependencies in the background. It speaks up when it has something worth saying — and stays quiet the rest of the time.

---

## Install

```bash
npm install -g @johndansu/flint
```

Or download a binary directly from [Releases](https://github.com/johndansu/Flint/releases).

Set your Anthropic API key:

```bash
export ANTHROPIC_API_KEY=sk-ant-...
```

---

## Quick start

```bash
flint init        # choose awareness level, set API key, install pre-commit hook
flint scan        # index your codebase and establish baselines
flintd            # start the background daemon (Flame / Forge mode)
```

That's it. Flint runs in the background from here.

---

## Awareness levels

| Level | What it does |
|-------|-------------|
| `spark` | Manual tools only. No background process. |
| `flame` | Daemon watches the current project. Observations fire automatically. |
| `forge` | Daemon watches all projects. |

Set at `flint init` or change anytime in `.flintrc`:

```json
{ "awareness": "flame" }
```

---

## CLI commands

```
flint init                     Set up Flint for this workspace
flint status                   Show daemon and config status
flint scan                     Index codebase, establish baselines

flint explain <file> [fn]      What this file/function does and why
flint diff [ref]               Plain-English explanation of a diff
flint watch "payment logic"    Set a tripwire — fires immediately when touched
flint watch list               List active tripwires
flint watch remove <id>        Remove a tripwire
flint fix --auto <category>    Enable auto-fix for a category (cve | import | format)
flint update [--format slack]  Last N commits as a stakeholder update
```

Every auto-fix is staged, never committed. You always review before it enters version control.

---

## Manual tools

Invoke any tool directly without the daemon:

```bash
flint <role> <category>
```

**Roles:** `webdev` · `fullstack` · `devops` · `cyber` · `data` · `mobile`

**Categories:** `doc` · `review` · `debug` · `audit`

```bash
flint webdev review       # Review recent changes as a web developer
flint devops audit        # Audit infrastructure and deployment config
flint cyber debug         # Debug a security issue
flint data review         # Review a data pipeline or model
```

Flint pulls your staged diff automatically. If there's nothing staged, it asks what to look at.

---

## VS Code extension

Install from the [VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=flintlang.flint) or from a `.vsix` in [Releases](https://github.com/johndansu/Flint/releases).

The extension connects to the local daemon over a Unix socket. When the daemon fires an observation, a card appears in the Flint sidebar — no popups, no interruptions. Dismiss when you're done.

Works in VS Code, Cursor, and Windsurf.

---

## How it works

Flint runs as a lightweight local daemon (`flintd`). It watches file saves, git activity, and dependency changes. When it notices something — after a period of coding stillness — it fires one observation. Not a stream. Not a checklist. One thing.

**Time gates** prevent noise:
- Human session: observation after 90 seconds of stillness, max 5 per session
- Vibe coding session: 10 seconds of stillness, max 8 per session (Flint detects the difference automatically)

**Footgun detection** runs before any AI call — 72 stack-aware patterns covering React, Go, Docker, Terraform, Python, and more. No API cost for known bad patterns.

**Tripwires** bypass all time gates. Set one on any code pattern and Flint fires the moment that code is touched.

All data stays local. SQLite at `~/.flint/flint.db`. No telemetry.

---

## Configuration

| File | Scope | Purpose |
|------|-------|---------|
| `~/.flint/config.json` | Global | API key, default awareness, model |
| `.flintrc` | Workspace | Per-project overrides |
| `ANTHROPIC_API_KEY` | Environment | API key fallback |

Workspace settings override global. Environment variable overrides both.

---

## Human intelligence layer

Off by default. Opt in at `flint init` or in config:

```json
{ "humanLayer": true }
```

When enabled, Flint watches for: debugging spirals, burnout signals, lone-wolf patterns, and interruption cost. Observations are infrequent and always framed as questions, not diagnoses.

---

## License

MIT
