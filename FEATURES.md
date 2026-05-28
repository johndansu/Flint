# Feature Registry
## Flint — Complete Feature List by Version

**Version:** 1.0
**Last updated:** May 2026
**Purpose:** Single source of truth for all features across all versions

---

## v1 — Foundation + Intelligence

### Core toolkit (24 tools)
6 roles × 4 tools. Manual invocation. Streaming output. Senior dev tone. "Why this matters" on every output.

### Daemon — technical intelligence
Full spec in MVP.md and DETECTION.md. Watches file activity, git, dependencies, patterns. Time-gated. Session-aware.

### Daemon — human intelligence (opt-in)
Energy patterns, burnout signals, career blind spots, interruption cost, cognitive load, growth signals, lone wolf, second system, impostor syndrome patterns.

### Session type detection (DETECTION.md)
12-signal multi-factor classifier. Human, agent-assisted, vibe coding, pair programming, mixed. Personal baseline builds over 10+ sessions. Agent fingerprinting for known agents.

### Vibe coding mode
Agent chunk completion detection. Acceptance risk monitoring. Agent prompt quality feedback. Adjusted time gates.

### Stack-aware footgun library
`shared/footguns.json` — role and framework specific. Covers: React, Vue, Node, Django, FastAPI, Go, Rust, Swift, Kotlin, React Native, Terraform, Docker. Categories: performance, security, concurrency, testing, deployment. Community-contributed, versioned by framework version.

### Onboarding scan — `flint scan`
Indexes existing codebase on first install. Establishes baselines. Writes to SQLite. Tells Flint "this is normal for this project." Required before daemon makes meaningful observations.

### Explainability
Every observation below 0.82 confidence includes a one-line "why I noticed this." Built into observation schema. Zero extra API cost.

### Flint Explain — `flint explain`
Reads a function or file. Explains what it does, why it exists, what assumptions it makes, what would break if changed. Comprehension assistance, not code review.

### Flint Diff — `flint diff`
Reads a git diff. Produces plain-English explanation of what changed, the apparent intent, and risks introduced. Works on staged changes, commit ranges, or PR diffs.

### Flint Watch — tripwires
Developer sets a tripwire: `flint watch "payment processing"`. Daemon monitors for any change touching that pattern and fires immediately — bypassing normal time gates. Managed via `flint watch list` and `flint watch remove`.

### Pre-commit checkpoint
Consequence analysis on every commit. Downstream impact, untested paths, injection risks. Never blocks — always advisory.

### Universal error log — `~/.flint/errors.log`
Every error from every source logged automatically. Terminal stderr, test failures, build errors, runtime crashes, pre-commit failures, CVE findings, install failures. All captured, all timestamped, all in one place. Human-readable log file mirrors the SQLite store. Grows silently. Never requires action. Becomes the most complete record of a developer's error history that any tool has ever built.

### Silent fix — `flint fix`
Two modes. Explicit: `flint fix <error-id>` — fix a specific logged error on demand. Auto-fix: opt-in by category (`flint fix --auto cve`, `--auto import`, `--auto format`). Every auto-fix is staged, never committed. Developer always reviews before it enters version control. Every fix logged in `auto_fixes` table. Flint always shows a card when an auto-fix fires — nothing is ever truly silent to the developer.

### Stakeholder translation — `flint update`
Last N commits in plain English. Format options: slack, email, plain. Audience options: dev, pm, cto.

### Flint-initiates communication model
Observation cards with brief follow-up window. Six card types. One follow-up per observation maximum. Then closes.

### CLI and extension independence
Full CLI without extension. Manual tools panel without CLI. Better together but fully functional apart. Cross-promotion contextual not timer-based.

### Multi-workspace isolation
Per-workspace daemon socket. Per-workspace config via `.flintrc`. CLI resolution in 5-step order.

### Extension operating modes
Connected, reconnecting, standalone, degraded. Manual tools available in all modes.

### Awareness levels
Spark (manual only), Flame (current project), Forge (everything).

---

## v2 — Workflow Intelligence + Stakeholder Layer

### Habit feedback
Commit size trending, session gaps, repeated error categories, consistency patterns. Long cooldowns. Never nagging.

### Daily brief — `flint brief`
State of codebase. Riskiest files. What you left half-done. Yesterday's patterns.

### Tool personalisation
Usage-based tool reordering per role. Most-used tools surface to top.

### Flint REPL — `flint repl`
Session-based terminal mode. Codebase context stays loaded. Brief focused sessions. Time-limited. Not an infinite chatbot.

### `flint pr` — PR review
Reads PR diff plus codebase context. Senior-dev quality review. Intent analysis. Risk assessment. Needs GitHub/GitLab/Bitbucket API integration.

### Dead code detector
Tracks function and module call frequency over time. Flags code untouched and unreferenced for 8+ months. "You might not need this anymore."

### Environment parity checker
Compares `.env`, `docker-compose.yml`, deployment configs. Surfaces "your local environment has 3 variables not in production."

### Dependency health scorecard
Beyond CVEs. Last updated, open issues, download trend, maintenance status, maintainer count, license compatibility. Weekly refresh. Proactive observation when health score drops.

### Pair programming awareness
Full detection from DETECTION.md. Adjusted observation threshold. Human intelligence suppressed. Knowledge transfer flag set.

### Flint Journal — `flint journal`
Auto-generated dev log from session activity. What you worked on, problems hit, problems solved, actual time spent. Weekly and daily views.

### Public observation library (beta)
Community-contributed anonymised observation patterns. Opt-in sharing. Curation process. Sync via `flint update-library`.

### Editor support expansion
JetBrains (one plugin, all IDEs). Neovim (Lua). Antigravity (VS Code fork, post-2.0 stability).

### Notification channels
Slack (engineering channels + DMs). Email (weekly digest). Telegram (scheduled stakeholder updates).

### Web dashboard
Next.js, localhost, stakeholder-facing. REST bridge at `localhost:7432`. Read-only from SQLite store.

---

## v3 — Codebase Intelligence + Expanded Reach

### Codebase indexer
tree-sitter AST parsing. Incremental updates. Dependency graph. Relevance-based context selection for API calls (max 50K tokens, never full codebase).

### Deep codebase learning
Stack recognition, naming convention learning, pattern memory, recurring mistake tracking. Advice sounds like it came from a 6-month team member.

### Cross-project memory
Re-onboarding brief when returning to a project after months away. "Here's what this does, where you left off, what changed in the ecosystem."

### Outcome feedback loop
Correlates past observations with future bugs. Weights observation categories based on whether ignored observations led to real problems. Requires codebase graph.

### Real-time doc generation
Watches what you build, drafts docs incrementally. 80% written by the time you finish coding.

### Systemic security patterns
Pattern-level security across the codebase. "You handle auth three different ways and two are inconsistent." Not line-level lint.

### Architectural opinion capability
With full codebase graph and history — opinion on structural decisions, not just code quality.

### API contract monitoring
Watches API responses against historical baselines. Fires immediately when a contract breaks.

### Flint Teach — personalised learning
Based on personal history — specific learning recommendation once per week. "You keep hitting async timing issues — here's the exact concept worth 30 minutes."

### Flint Mood — codebase health trend
Weekly codebase health direction. Tests added or removed. Debt accelerating or decelerating. Documentation improving. Numbers not vibes.

### Agent quality differentiation
Over time, learns which agent produces code that holds up better in this codebase. Surfaces as a gentle observation.

### Offline / local model fallback
Local inference for core observations. Ollama integration. Reduced quality but functional in air-gapped environments.

### Notification channels expansion
WhatsApp (Business API — start approval process in v2 timeline). Discord (webhook). Notion (API integration). Linear (ticket integration).

### Editor support expansion
Zed. Others as ecosystem matures.

---

## v4 — Team Intelligence

### Team pattern aggregation
Systemic issues surfaced at team level. Never individual. "Your team hits this error category 3x more than comparable teams."

### Knowledge distribution alerts
Single points of knowledge on critical codebase areas. Surfaces before that person leaves.

### Git blame intelligence
Contextual code history. "This was written under deadline pressure — worth a closer look." Never for blame.

### Shared runbooks
Get smarter from team usage. Deviations become improvement suggestions.

### Team stakeholder reports
Weekly engineering digest. Auto-generated from git activity and codebase health. Zero manual effort.

### Encrypted P2P team sync
No central server. Individual data never shared. Only aggregated patterns.

### Team human intelligence
Burnout signals at team level (never individual). Aggregate only. Opt-in at team level.

### Flint for code review — async layer
PR review extended to team workflow. Which PRs take longest. Which areas get most review comments. Knowledge distribution via review patterns.

### Public observation library (full)
Full community curation. Agent fingerprint contributions. Stack footgun contributions. Versioned by framework.

### License and team seat management
Offline-capable license validation. Ed25519 signatures.

---

## Feature Count

| Version | New features | Cumulative |
|---|---|---|
| v1 | 24 | 24 |
| v2 | 17 | 41 |
| v3 | 14 | 55 |
| v4 | 12 | 67 |

---
