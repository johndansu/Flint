# Changelog
## Flint

All notable changes to Flint will be documented here.
Format: [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)

---

## [Unreleased] — v1.0.0

### Added

**Core toolkit**
- CLI tool (`flint`) distributed via npm wrapper over Go binary
- VS Code, Cursor, and Windsurf extension with sidebar presence
- Three awareness levels: Spark, Flame, Forge
- Human intelligence layer (opt-in, `flint human on`)
- 24 role-aware manual tools across 6 roles

**Session intelligence**
- 12-signal multi-factor session type classifier
- Session types: human, agent-assisted, vibe coding, pair programming, mixed
- Personal developer baseline (builds over 10+ sessions)
- Agent fingerprinting for known agents (Cursor, Antigravity, Windsurf)
- Vibe coding mode with agent chunk evaluation and acceptance risk monitoring
- Per-session type time gate adjustments

**Daemon — technical intelligence**
- File watcher, git monitor, dependency CVE checker
- Stack-aware footgun library (`shared/footguns.json`) — React, Go, Django, Node, and more
- Onboarding scan (`flint scan`) — indexes codebase, establishes baselines before daemon runs
- Explainability — "why I noticed this" on observations below 0.82 confidence
- Observation categories: code quality, git behaviour, dependencies, test coverage, codebase health, consistency, wins, taste, questions, deliberate silence
- Tripwire engine (`flint watch`) — bypasses time gates, fires immediately on pattern match

**Daemon — human intelligence (opt-in)**
- Energy and focus patterns, burnout early signals, career blind spots
- Interruption cost, cognitive load, growth vs stagnation
- Lone wolf signal, second system signal, impostor syndrome patterns
- 0.85 confidence threshold, 7-day cooldown, 2-week minimum observation window

**New manual tools + REPL**
- `flint explain` — comprehension assistance for unfamiliar code
- `flint diff` — plain-English diff explanation with risk analysis
- `flint repl` — interactive codebase session. Loads context once, maintains conversation in memory, clears on close. No persistence between sessions. Not a chatbot — a focused depth tool.
- `flint watch` / `flint watch list` / `flint watch remove` — tripwires

**Communication model**
- Flint-initiates model with brief follow-up window
- Six card types with different action sets
- One follow-up per observation maximum, then closes
- Agent prompt generation formatted per connected agent

**CLI + Extension independence**
- Full CLI without extension (manual tools + pre-commit + terminal workflows)
- Extension standalone mode (Spark-level manual tools, no daemon required)
- 5-step CLI resolution order in extension
- Per-workspace daemon socket isolation
- `.flintrc` per-project config with team commit support
- Four extension operating modes: connected, reconnecting, standalone, degraded
- Contextual cross-promotion (never timer-based, suppressible)

**Error log + auto-fix**
- Universal error logger — captures all 7 sources automatically (terminal, test, build, runtime, precommit, dependency, install)
- Human-readable `~/.flint/errors.log` — grows silently, always local
- Error signature computation and status tracking (observed/fixed/ignored/auto_fixed)
- Auto-fix engine — staged-only, never commits, always notifies
- CVE auto-fix (`flint fix --auto cve`)
- Import auto-fix (`flint fix --auto import`)
- Format auto-fix (`flint fix --auto format`)
- `flint errors` command with full filter options
- `flint fix` command — manual and auto-fix configuration
- Auto-fix history log (`flint fix --log`)
- Go + TypeScript monorepo with shared JSON schema contract
- Cross-compiled binaries: macOS arm64/amd64, Linux amd64, Windows amd64
- GitHub Actions CI/CD pipeline
- SQLite local store with 6 tables
- Unix socket IPC between daemon and extension

---

## Planned — v2.0.0

### To Add
- Habit feedback and pattern detection
- Daily brief (`flint brief`)
- Tool personalisation (usage-based reordering)
- JetBrains plugin (IntelliJ, WebStorm, PyCharm, GoLand, Rider)
- Neovim plugin
- Antigravity extension
- Slack notification channel
- Email digest
- Telegram bot (stakeholder broadcast)
- Web dashboard (Next.js, localhost, stakeholder-facing)
- Local REST bridge (`localhost:7432`)
- Deeper human pattern tracking (longer observation windows)
- Role auto-detection from project files

---

## Planned — v3.0.0

### To Add
- Codebase indexer (tree-sitter AST parsing)
- Dependency graph builder
- Cross-project memory store
- Real-time documentation generation
- Systemic security pattern detection
- Architectural opinion capability
- WhatsApp Business API integration
- Discord webhook integration
- Notion API integration
- Linear ticket integration
- Zed editor plugin
- Context window strategy for large codebases

---

## Planned — v4.0.0

### To Add
- Team pattern aggregation (never individual)
- Knowledge distribution alerts
- Shared runbooks
- Team stakeholder reports
- Encrypted P2P team sync
- License and team seat management
- Team-level human intelligence (aggregate only)

---
