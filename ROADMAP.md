# Flint — Build Roadmap
## A Systematic Guide to Building Each Version

**Version:** 1.0
**Last updated:** May 2026
**Purpose:** Step-by-step build order for every version — what to build, in what order, and why that order matters

---

> Read this before writing a single line of code.
> The order here is not arbitrary. Every phase builds the foundation the next phase needs.
> Skipping ahead breaks things. Following this order means nothing is ever rebuilt.

---

## How to Read This Document

Each phase has:
- **What you're building** — the specific components
- **Why this order** — what breaks if you do it later
- **Done when** — the exact condition that tells you the phase is complete
- **Do not proceed until** — the hard gate before the next phase

---

# VERSION 1

---

## Phase 1 — The Contract
**Duration: 3–5 days**

Before any code. Before any features. Before anything.

### What you're building

**`shared/schema.json`** — every IPC message type between Go and TypeScript. This is the contract. Everything else is generated from it.

**`shared/tools.json`** — all 24 tool definitions. Role, category, system prompt, placeholder, hint. Written once, used everywhere.

**`shared/footguns.json`** — the complete footgun library. 61 entries, all fields populated including `detect_pattern` and `fix_confidence`.

**`scripts/generate-schema.js`** — the codegen script that reads `schema.json` and outputs:
- `core/pkg/schema/schema.go` — Go structs
- `extension/src/ipc/messages.ts` — TypeScript interfaces

### Why this order

If you build Go code before the schema, you will write your own message types and they will drift from TypeScript. If you write the schema after building, you are retrofitting a contract onto code that already made assumptions. The schema must exist before either side touches IPC.

The tool prompts must be written before building the AI integration — otherwise you will hardcode prompts into Go files and spend a week extracting them later.

The footgun library must be written before the daemon — the daemon needs to know what patterns to look for.

### Done when
- `shared/schema.json` validates with all message types defined
- `shared/tools.json` has all 24 tools with complete prompts tested manually against the API
- `shared/footguns.json` has all 61 entries with `detect_pattern` fields
- `node scripts/generate-schema.js` runs without errors and produces both output files
- Both generated files compile without errors (`go build` passes, `tsc` passes)

### Do not proceed until
Generated types compile cleanly on both sides with zero errors.

---

## Phase 2 — Go Core Foundation
**Duration: 5–7 days**

The shared Go infrastructure everything else depends on.

### What you're building

**`core/internal/config/config.go`**
- Read/write `~/.flint/config.json`
- Fields: role, awareness level, API key, agent, human layer toggle, auto-fix settings, nudge suppression
- Creates `~/.flint/` directory on first run if it doesn't exist

**`core/internal/store/store.go`** + all table files
- SQLite wrapper using `modernc.org/sqlite`
- All 9 tables: observations, error_log, error_dictionary, auto_fixes, calibration, sessions, human_signals, timegates, tool_usage
- WAL mode enabled
- Migration system for future schema changes

**`core/internal/ai/client.go`**
- Anthropic API HTTP client
- Streaming response handler — tokens to `io.Writer`
- Retry logic with exponential backoff
- Error handling for rate limits, auth failures, timeouts

**`core/internal/awareness/awareness.go`**
- Spark / Flame / Forge constants
- Gate functions: `CanUseManualTools()`, `CanRunDaemon()`, `CanTrackHuman()`

**`core/internal/tools/tools.go`**
- Load `shared/tools.json` at startup
- `Get(toolID, role) Tool` — returns correct tool for role
- `ListForRole(role) []Tool` — all tools for a role

### Why this order

Config must exist before anything else reads it. Store must exist before anything writes observations, errors, or sessions. AI client must exist before any tool runs. Tools must load from JSON before the CLI can invoke them. Awareness gates must exist before the daemon respects them.

### Done when
- `go test ./internal/config/...` passes
- `go test ./internal/store/...` passes with all tables created and queried
- AI client streams a real response from the Anthropic API
- `tools.Get("review", "webdev")` returns the correct prompt from `tools.json`

### Do not proceed until
All four packages have passing tests and the AI client has streamed at least one real response.

---

## Phase 3 — CLI Foundation (Spark Level)
**Duration: 5–7 days**

The CLI that works without a daemon. Manual tools only. This is what developers install first.

### What you're building

**`core/cmd/flint/main.go`** — cobra root command + all subcommands registered

**`flint init`**
- First run wizard: role selection, awareness level, API key, agent selection, human layer opt-in
- Writes `~/.flint/config.json`
- Installs git pre-commit hook if awareness >= Flame
- Prints welcome message

**`flint review`, `flint debug`, `flint doc`, `flint audit`**
- Read from stdin (`io.ReadAll(os.Stdin)`)
- Load tool from `tools.json` for configured role
- Call AI client with streaming
- Stream to stdout
- After output: check nudge engine, maybe print extension nudge

**`flint explain`, `flint diff`**
- Same pattern as above, different prompts

**`flint config`, `flint awareness`, `flint status`**
- Config management
- Status shows: awareness level, role, agent, daemon state (stopped)

**`flint memory`** + `--clear` + `--export`
- Read from SQLite store, print formatted
- Clear deletes all tables
- Export writes JSON to stdout

**`core/internal/nudge/nudge.go`**
- Tool usage counter (per tool, per day in `tool_usage` table)
- 3-nudge-per-day threshold
- Print nudge after output, never interrupt
- `flint config --no-nudge` suppression

**`wrapper/package.json`** + `wrapper/install.js` + `wrapper/bin/flint.js`
- Platform detection (darwin-arm64, darwin-amd64, linux-amd64, win32-x64)
- Binary download from GitHub Releases on install
- Shim that exec's the Go binary

### Why this order

The CLI is the foundation. It must work completely at Spark level before the daemon is built — because the daemon depends on the same config, store, and tools that the CLI uses. Building the CLI first validates all the foundation code under real use.

The wrapper is built now because it is how you distribute the CLI for testing. You cannot give 20 developers a raw Go binary to test — you give them `npm install -g flint`.

### Done when
- `npm install -g flint` works on macOS arm64 and Linux amd64
- `flint init` completes the wizard and writes config
- `cat some_file.js | flint review` streams a real code review
- `cat error.log | flint debug` streams a real error explanation
- `flint status` shows current config correctly
- `flint memory --clear` deletes all data
- A nudge appears after 3+ tool uses in a day

### Do not proceed until
The CLI is installed via npm on at least two different machines and all manual tools produce good output.

---

## Phase 4 — Error Log + Auto-Fix
**Duration: 4–5 days**

The error log is always on. Build it before the daemon so it captures errors from CLI tool runs too.

### What you're building

**`core/internal/errorlog/logger.go`**
- Universal error logger — captures from all 7 sources
- Signature computation (hash of error type + stack pattern)
- Status tracking (observed / fixed / ignored / auto_fixed)
- Writes to `error_log` SQLite table
- Mirrors to `~/.flint/errors.log` human-readable file

**`core/internal/errorlog/autofix.go`**
- Auto-fix engine
- CVE auto-fix: runs `npm audit fix` or equivalent, stages result
- Import auto-fix: resolves unambiguous missing imports, stages result
- Format auto-fix: runs project formatter, stages result
- **Staged-only constraint** — `git add` but never `git commit`
- Always broadcasts notification card to daemon (if running) or prints to terminal

**`flint errors`** command + all filters
**`flint fix`** command + `--auto` options + `--log`

**`core/internal/errorlog/sources/`**
- Terminal stderr watcher (captures from shell integration)
- Test runner parser (Jest, pytest, go test output formats)
- Build failure parser (TypeScript, webpack, go build output)

### Why this order

Error logging must exist before the daemon starts — the daemon will add more error sources but the core logger needs to be solid first. Auto-fix must be built and tested before the daemon runs it automatically — the "staged only, never commits" constraint is critical and needs to be verified manually before it runs silently.

### Done when
- Running a failing test logs the error to `~/.flint/errors.log`
- `flint errors` shows recent errors with correct status
- `flint fix --auto cve` updates a vulnerable dependency and stages the change
- `git diff --staged` shows exactly what was changed
- `git status` shows no automatic commits were made

### Do not proceed until
Manually verify on a real project: auto-fix stages a change, never commits it, and the notification prints clearly.

---

## Phase 5 — Pre-Commit Checkpoint + `flint scan`
**Duration: 4–5 days**

Two foundational features that work without the passive daemon.

### What you're building

**`core/internal/daemon/scanner.go`** — `flint scan`
- Walk entire codebase
- Count files, functions, dependencies
- Detect stack from `package.json`, `requirements.txt`, `go.mod`, etc.
- Load matching footguns from `shared/footguns.json`
- Establish baselines: avg function length, test coverage ratio, dependency health
- Write baselines to SQLite store
- Print scan summary

**`core/internal/daemon/precommit.go`** — `flint check`
- Read `git diff --staged`
- Map changed files to downstream dependents
- Check changed code against footgun library patterns
- Call AI for consequence analysis
- Print structured report
- Return exit code 0 (always advisory, never blocking)

**`core/cmd/flint/repl.go`** — `flint repl`
- Load codebase context from scan baseline on session open
- Maintain messages array in memory for session duration
- Stream responses with full context
- Hard context window limit (drop oldest exchange at 20 messages)
- `exit` / Ctrl+C closes session, clears all memory
- No persistence between sessions — always starts fresh
- Session header shows project path, awareness level, context loaded

**Git hook installer** in `flint init`
- Write `.git/hooks/pre-commit` that calls `flint check`
- Make executable
- `flint init --no-hook` to skip

### Why this order

`flint scan` establishes the baselines the daemon uses. The daemon should not start making observations on a codebase it has never scanned — observations will be noisy without a baseline. `flint check` is built now because it is the most immediately impressive feature and works without a running daemon. It validates the core consequence-analysis AI prompt before the daemon uses similar logic.

### Done when
- `flint scan` on a real project completes and writes baselines to SQLite
- `flint scan` correctly identifies the stack (React, Go, Python, etc.)
- `git commit` on a project with the hook installed runs `flint check` and shows a real consequence report
- `flint check` catches a downstream impact that the developer didn't notice

### Do not proceed until
`flint check` has caught something real on a real codebase.

---

## Phase 6 — The Daemon Core
**Duration: 7–10 days**

The background process. The heart of Flint.

### What you're building

**`core/internal/daemon/daemon.go`** — daemon lifecycle
- Start / stop / PID management
- `~/.flint/daemon-{workspace-hash}.sock` Unix socket server
- Broadcast to all connected clients

**`core/internal/daemon/watcher.go`** — file system watcher
- fsnotify integration
- Debounce rapid changes
- Track file change history per session

**`core/internal/daemon/git.go`** — git monitor
- Periodic `git log`, `git diff` polling
- Commit frequency tracking
- Commit message quality scoring

**`core/internal/daemon/deps.go`** — dependency monitor
- CVE database fetch and local cache (weekly refresh)
- Health score tracking (maintenance, download trend)
- Immediate fire on CVE detection (bypasses all time gates)

**`core/internal/daemon/timegate.go`** — time-gating rules
- 20-minute minimum gap between observations
- 5 observations max per session
- 24-hour category cooldown
- 15-minute session silence at start
- Per-session type overrides (vibe: 10 min, 8 max)

**`core/internal/daemon/technical.go`** — technical observation engine
- Code quality checks (post-stillness only)
- Git and commit behaviour checks
- Test coverage checks
- Consistency checks
- Footgun pattern matching (fast regex first, AI second)
- Win signal detection
- Taste observations
- Intent questions

**`flint start`, `flint stop`, `flint status`** — daemon lifecycle CLI commands

### Why this order

The daemon is built after the foundational tools because it reuses the config, store, AI client, and error logger already built. The time gate is built before the observation engines — otherwise you will add time gates as an afterthought and they will be inconsistent. The technical engine is built before the session classifier because you need to see what observations look like before you start tuning when they fire.

### Done when
- `flint start` launches the daemon process
- `flint stop` kills it cleanly
- `flint status` shows daemon running with correct project path
- The daemon detects a CVE and the notification appears in the terminal (as a printed line since extension isn't built yet)
- The daemon makes at least one technical observation on a real codebase without being asked

### Do not proceed until
The daemon runs for 30 minutes on a real project without crashing, consuming > 1% CPU, or using > 50MB RAM.

---

## Phase 7 — Session Type Classifier
**Duration: 5–7 days**

Now the daemon can tell who is writing.

### What you're building

**`core/internal/daemon/classifier.go`** — 12-signal classifier
- All 12 signal functions
- Weighted confidence vector
- `SessionTimeline` with transition logging
- `TypeAtTime(t time.Time) SessionType` for observation attribution

**`core/internal/daemon/baseline.go`** — personal baseline
- Builds over first 10+ confirmed-human sessions
- Per-signal baseline values and standard deviations
- Stores in SQLite, persists across restarts

**`core/internal/daemon/fingerprint.go`** — agent fingerprinting
- Load `shared/agent_fingerprints.json`
- Score agent chunks against known fingerprints

**Vibe coding mode** in `technical.go`
- Agent chunk completion detection (10s stillness)
- Agent output evaluator
- Acceptance risk detector (< 30s review window)
- Agent prompt quality observer
- Adjusted time gates

### Why this order

The classifier is built after the basic daemon because you need real session data to tune it. Running the basic daemon for a few days first gives you real keystroke patterns, change sizes, and session behaviour to validate the classifier against. Building the classifier on synthetic data produces a classifier that doesn't work on real sessions.

### Done when
- Classifier correctly identifies your own coding session as `human` >= 90% of the time
- Classifier identifies a Cursor agent session as `agent_assisted` or `vibe_coding`
- `flint status` shows current session type
- Vibe coding acceptance risk fires on a real fast-accept scenario

### Do not proceed until
Test the classifier on 5 different session types and verify accuracy before proceeding.

---

## Phase 8 — Human Intelligence Layer
**Duration: 5–7 days**

The layer that cares about the developer, not just the code.

### What you're building

**`core/internal/daemon/human.go`** — human observation engine
- All 9 human signal trackers:
  1. Session length and frequency trends
  2. Commit quality distribution by time of day
  3. Revert rate trend monitor
  4. Interruption detection and cost estimator
  5. Career activity distribution
  6. Cognitive load indicator
  7. Growth vs stagnation signal
  8. Lone wolf detector
  9. Second system signal
  10. Impostor syndrome patterns (Forge, 4+ weeks, 0.85 threshold)

- 0.85 confidence threshold (higher than technical)
- 7-day cooldown per category
- Maximum 1 human observation per session
- Minimum 2-week observation window before first observation

**`flint human on/off/status/--clear`** commands

### Why this order

Human intelligence requires session history to be meaningful. Building it after 3+ weeks of using the daemon means you already have real session data to test against. Building it earlier means testing against synthetic data that doesn't represent real behaviour.

The 2-week minimum observation window is not just a feature — it is a quality gate. The first 2 weeks of data are noisy. Human observations should never fire before the window.

### Done when
- `flint human on` enables the layer
- Session length data accumulates correctly in `human_signals` table
- After 2+ weeks of real usage: at least one human observation fires that feels accurate
- `flint human --clear` deletes all human data without affecting technical data

### Do not proceed until
At least one human observation fires naturally during real usage and feels genuinely useful rather than obvious.

---

## Phase 9 — VS Code + Cursor + Windsurf Extension
**Duration: 7–10 days**

The editor surface. One build, three marketplaces.

### What you're building

**`extension/src/cli/resolver.ts`** — 5-step CLI resolution
**`extension/src/ipc/socket.ts`** — per-workspace socket path
**`extension/src/ipc/client.ts`** — Unix socket client with reconnection
**`extension/src/mode.ts`** — connected / reconnecting / standalone / degraded
**`extension/src/status/bar.ts`** — all four mode states

**`extension/src/sidebar/provider.ts`** — sidebar webview
- Quiet state: 4px accent strip
- Speaking state: observation card
- Card type router (6 types)
- Technical card: observation + agent prompt + Copy + Send + Tell me more + ✕
- Follow-up window: streams one response, then locks
- Win card: fades 5s, no actions
- Human card: fades 8s, no actions, no ✕
- Intent question card: Tell me more, no ✕
- Taste card: Tell me more + ✕

**`extension/src/agent/`** — agent send integrations
- Cursor, Antigravity, Windsurf, Copilot, Custom

**`extension/src/nudge/cross-promotion.ts`** — contextual nudges

**`extension/package.json`** configured for VS Code, Cursor, and Windsurf marketplaces

### Why this order

The extension is built last among v1 components because it depends on everything else. The daemon must be running and broadcasting observations before the extension has anything to display. The CLI resolution must work before the extension knows which daemon to connect to.

Building the extension before the daemon means building against a mock that will inevitably diverge from reality.

### Done when
- Extension installs from marketplace in VS Code
- Quiet state (4px strip) appears when extension activates
- A daemon observation triggers a card in the sidebar
- "Tell me more" streams a follow-up and locks
- "Send to Cursor" correctly passes the prompt to Cursor
- Standalone mode shows manual tools panel with no daemon
- Status bar shows correct mode in all four states
- Contextual nudge appears when daemon is absent for 10+ active minutes

### Do not proceed until
End-to-end test: write code in VS Code, let the daemon observe, see a card appear in the sidebar, tap "Tell me more", verify the follow-up is useful.

---

## Phase 10 — Forge Features
**Duration: 3–4 days**

The features that require history depth.

### What you're building

**Personal error dictionary** — built from `error_log` over time
- Signature matching on `flint debug` runs
- "You've seen this before" card in sidebar
- Same card in terminal output

**Stakeholder translation** — `flint update`
- `git log --oneline -N` + diff summaries
- AI translation with audience-tuned prompt
- Format options: plain, slack, email
- Audience options: dev, pm, cto

**Cross-project pushback** — in `technical.go`
- Pattern matching across `error_log` entries from different projects
- Explicit cross-project callout in observation text

**`flint watch`** — tripwire engine
- `flint watch "pattern"` stores tripwire in SQLite
- Daemon checks all tripwires on every file change
- Bypasses all time gates when triggered
- `flint watch list`, `flint watch remove <id>`

### Done when
- `flint update` produces a stakeholder-ready summary from real git history
- Error dictionary matches a real repeated error and surfaces the previous fix
- A tripwire fires correctly and immediately when the watched pattern is touched

---

## Phase 11 — Internal Testing + Threshold Tuning
**Duration: 10–14 days**

The most important phase. Do not rush it.

### What you're doing

Use Flint on real projects every day. Take notes on everything:

**For each daemon observation that fires:**
- Was it worth the interruption?
- Was the timing right?
- Did the agent prompt fix the issue?
- Was the follow-up useful or did it repeat the observation?

**For vibe coding sessions:**
- Did the classifier correctly identify the session type?
- Did acceptance risk fire at the right moment?
- Was the "three things to check" list accurate?

**For human intelligence observations:**
- Did they feel accurate or presumptuous?
- Was the timing right (not too early in a session)?
- Was the language right (not clinical, not obvious)?

**Tune based on findings:**
- Adjust confidence thresholds per category
- Adjust time gates if observations feel too frequent or too rare
- Rewrite observation prompts that produce generic output
- Fix footgun patterns that produce false positives

**Target state before soft launch:**
- Daemon fires 1–3 observations per 4-hour session
- Every observation feels worth pausing for
- No observation has fired twice for the same issue in the same session
- Pre-commit checkpoint catches something real at least once per week
- Auto-fix has staged at least 3 CVE fixes correctly

### Done when
All targets above are met on at least 2 different real projects over at least 10 days of use.

### Do not proceed until
You personally feel that Flint is something you would miss if it were removed.

---

## Phase 12 — Build Pipeline + Distribution
**Duration: 3–4 days**

Everything needed to ship to real users.

### What you're building

**`Makefile`** — all build targets
**`scripts/cross-compile.sh`** — all 4 platform binaries
**`.github/workflows/ci.yml`** — test on every PR
**`.github/workflows/release.yml`** — cross-compile + publish on tag push

**Release process:**
```bash
git tag v1.0.0-beta.1
git push origin v1.0.0-beta.1
# → CI builds all binaries
# → uploads to GitHub Releases
# → publishes npm wrapper
# → publishes VS Code extension (to VS Code Marketplace, Cursor, Windsurf)
```

### Done when
- `npm install -g flint` works on macOS arm64, macOS amd64, Linux amd64, Windows amd64
- Extension installs from VS Code Marketplace
- A release tag triggers the full pipeline automatically

---

## Phase 13 — Private Beta (20 Developers)
**Duration: 4–6 weeks**

### Who to recruit
- 5 web/full-stack developers
- 4 DevOps/cloud engineers
- 3 mobile developers
- 3 data/ML engineers
- 3 security engineers
- 2 solo founders / indie developers

Mix of: human coders, vibe coders, pair programmers. Different editors. Different OS.

### What to give them
- `npm install -g flint` + extension link
- A 5-minute setup guide (not documentation — a guide)
- Direct access to you (not a feedback form)
- Weekly 20-minute call optional

### What to measure
- Awareness level upgrade rate (Spark → Flame → Forge)
- Which tools get used most
- Which observations get dismissed most (calibration signal)
- Human layer opt-in rate
- Session type classifier accuracy (ask them)
- Whether anyone uses `flint watch`
- Whether `flint update` actually gets used for stakeholder comms

### The gate to public launch
Both must be true before proceeding:
1. 30-day retention above 40%
2. At least 5 unprompted "I would miss this if it were removed"

---

## Phase 14 — Public Free Beta
**Duration: Ongoing**

When the private beta gate passes:

1. **Hacker News Show HN** — lead with what it caught, not what it does
2. **VS Code Marketplace organic discovery** — good description, real screenshots
3. **npm page** — install count is social proof
4. **Developer Twitter/X** — one real observation screenshot per week
5. **The `flint update` share loop** — optional "generated with Flint" on stakeholder updates

Stay free. Gather signal. Plan v2.

---

# VERSION 2

> Build v2 only after the v1 private beta gate passes and you have real usage data.
> The order within v2 is driven by what the beta data tells you users want most.

---

## v2 Phase 1 — Workflow Intelligence (2–3 weeks)

Build in this order:

1. **Session history depth** — ensure 8+ weeks of session data in SQLite before building features that need it
2. **Habit feedback** — commit size trending, session gaps, repeated errors. Test against real session history.
3. **Daily brief (`flint brief`)** — state of codebase, riskiest files, what you left half-done
4. **Tool personalisation** — reorder tools based on usage data already collected

---

## v2 Phase 2 — `flint pr` + Dead Code Detector (2–3 weeks)

1. **`flint pr`** — GitHub/GitLab/Bitbucket API integration. Read PR diff + codebase context. Senior-dev review.
2. **Dead code detector** — function call frequency tracking. Requires 8+ weeks of daemon usage data to be meaningful.
3. **Environment parity checker** — compare .env, docker-compose, deployment configs

---

## v2 Phase 3 — Stakeholder Channels (2–3 weeks)

Build in this order — each validates before the next:

1. **Email digest** — simplest. SMTP/SendGrid. Weekly summary. No new infrastructure.
2. **Telegram bot** — Telegram Bot API. Schedule-based. Test with a small group first.
3. **Slack integration** — more complex. OAuth. Channel posting. Test with one team.
4. **Web dashboard** — Next.js, localhost, REST bridge. Build last — needs data from all channels.

---

## v2 Phase 4 — Editor Expansion (3–4 weeks)

1. **Cursor** — if not already identical to VS Code (it mostly is)
2. **Neovim plugin** — Lua. Connects to same daemon socket. Terminal-native audience will love this.
3. **JetBrains plugin** — Kotlin. One plugin covers all IDEs. Larger effort but large user base.
4. **Antigravity** — after 2.0 stability confirmed. VS Code fork so low effort.

---

# VERSION 3

> Build v3 only after v2 has been live for 3+ months with 500+ active users.
> The codebase indexer requires significant infrastructure. Don't build it speculatively.

---

## v3 Phase 1 — Codebase Indexer (4–6 weeks)

The hardest technical challenge in the entire roadmap.

1. **tree-sitter Go bindings** — parse ASTs for all supported languages
2. **Incremental indexer** — watch file changes, update index in place
3. **Dependency graph builder** — function/module call relationships
4. **Relevance selector** — given a file change, walk graph and select relevant subgraph (max 50K tokens)
5. **Index storage** — `~/.flint/index/{project-hash}/`
6. **`flint scan` upgrade** — now uses tree-sitter for precise baselines

---

## v3 Phase 2 — Deep Intelligence Features (4–6 weeks)

Each depends on the indexer:

1. **Deep codebase learning** — stack recognition, naming convention learning, pattern memory
2. **Cross-project memory** — re-onboarding brief using index diff
3. **Systemic security patterns** — pattern-level security using dependency graph
4. **Real-time doc generation** — incremental doc drafting using AST

---

## v3 Phase 3 — Outcome Feedback Loop (2–3 weeks)

Correlate past observations with future bugs:

1. **Observation → bug correlation** — when an error occurs, check if a related observation was dismissed
2. **Weight adjustment** — increase category weight if ignored observation preceded a real bug
3. **Transparency** — `flint memory --correlations` shows what Flint was right about that was ignored

---

## v3 Phase 4 — Expanded Channels + Offline (3–4 weeks)

1. **WhatsApp Business API** — start Meta approval process in v2 timeline, not v3
2. **Discord webhook** — simple, quick
3. **Notion/Linear integration** — more complex, requires API auth
4. **Offline/local model fallback** — Ollama integration, last because quality gap is real

---

# VERSION 4

> Build v4 only after v3 has been live for 3+ months with team adoption signals.
> v4 is a different product category — team intelligence. Don't rush it.

---

## v4 Phase 1 — Team Sync Infrastructure (4–6 weeks)

1. **Keypair generation** — Ed25519 per developer on `flint init`
2. **Peer discovery** — `flint team invite <email>`, exchange public keys
3. **Encrypted observation sharing** — individual data never shared, only patterns
4. **P2P sync** — direct device-to-device, no central server

---

## v4 Phase 2 — Team Intelligence Features (3–4 weeks)

Each depends on sync infrastructure:

1. **Team pattern aggregation** — surface systemic issues, never individual
2. **Knowledge distribution alerts** — single points of knowledge detection
3. **Shared runbooks** — team usage improves runbooks over time
4. **Git blame intelligence** — contextual code history for team context

---

## v4 Phase 3 — Monetisation Layer (2–3 weeks)

By v4, free beta has proven value. Now charge.

1. **License key system** — Ed25519 signed, offline-capable
2. **Seat management** — team seat count validation
3. **Grandfathering** — all v1 beta users get Forge free forever (non-negotiable)
4. **Billing integration** — Stripe or Paddle
5. **Pricing tiers** — Spark free, Flame $9/mo, Forge $18/mo, Team $15/seat/mo

---

## The One Rule That Governs All of This

> Never build the next phase until the current phase has been used on real code by a real developer and produced something genuinely useful.

Synthetic tests catch bugs. Real use catches wrong assumptions. Every phase gate exists to force real use before moving forward.

The roadmap is not a schedule. It is an order. Some phases will take longer than estimated. That is fine. What is not fine is skipping a phase because it feels complete before it has been used.

---

## Quick Reference — Phase Order

### v1
```
Phase 1  → The Contract (schema, tools, footguns)
Phase 2  → Go Core Foundation
Phase 3  → CLI Foundation (Spark) + flint repl
Phase 4  → Error Log + Auto-Fix
Phase 5  → Pre-Commit + flint scan
Phase 6  → Daemon Core
Phase 7  → Session Type Classifier
Phase 8  → Human Intelligence Layer
Phase 9  → VS Code + Cursor + Windsurf Extension
Phase 10 → Forge Features
Phase 11 → Internal Testing + Threshold Tuning ← most important
Phase 12 → Build Pipeline + Distribution
Phase 13 → Private Beta (20 developers)
Phase 14 → Public Free Beta
```

### v2
```
Phase 1  → Workflow Intelligence
Phase 2  → flint pr + Dead Code + Environment Parity
Phase 3  → Stakeholder Channels
Phase 4  → Editor Expansion
```

### v3
```
Phase 1  → Codebase Indexer
Phase 2  → Deep Intelligence Features
Phase 3  → Outcome Feedback Loop
Phase 4  → Expanded Channels + Offline
```

### v4
```
Phase 1  → Team Sync Infrastructure
Phase 2  → Team Intelligence Features
Phase 3  → Monetisation Layer
```

---
