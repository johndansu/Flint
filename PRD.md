# Product Requirements Document
## Flint — The Senior Dev That Never Leaves

**Version:** 0.7
**Status:** Draft
**Author:** TBD
**Last updated:** May 2026

---

## 1. Problem Statement

Developers have two kinds of problems.

The first kind they know about — bugs, errors, security issues, missing tests. Tools exist for these. They are well-served.

The second kind they don't know about — and these are the ones that actually define a career and break a codebase.

Context silently draining from their mental model of the codebase. Decision fatigue accumulating invisibly through the day. Knowledge silos forming without anyone deciding to form them. Technical debt accruing not in big choices but in six weeks of small ones. Documentation decaying the moment it's written. Burnout approaching long before it arrives. Career blind spots widening because nobody is watching what they're never doing.

These problems have no tools. They don't even have names most of the time. Developers feel them as vague friction — "this codebase is getting harder," "I've been less productive lately," "I don't really understand this part anymore" — without ever identifying the root cause.

**Flint is built for the second kind.**

Not just a better linter, not just a smarter code reviewer. A presence that understands developers as humans who have invisible problems — technical and human — and quietly helps with both.

---

## 2. Vision

> "Not a tool you use. A senior dev that never leaves — one who cares about your code and about you."

Flint is passive intelligence with a voice. It lives in your environment, watches everything you allow it to watch, and speaks up when staying silent would be the worse choice.

It operates on two planes simultaneously:

**Technical intelligence** — code quality, security, architecture signals, codebase health, knowledge silos, debt accrual rate, documentation decay.

**Human intelligence** — energy and focus patterns, burnout early signals, career blind spots, cognitive load trends, interruption cost, growth and stagnation.

Both planes observe the same thing — your work. The code is the data. The human is the context. The same voice, the same sidebar, the same tone. One presence, two depths.

---

## 3. The Design Principles

**Flint observes behaviour, never judges character.**
It notices patterns. It never makes diagnoses. It surfaces what it sees, never prescribes what the developer should feel.

Not: "you seem burned out."
But: "your sessions have been shorter and your revert rate is up — that sometimes means something's getting in the way. Just worth knowing."

The difference is everything.

**Earn every interruption.**
Silence is the default. Flint speaks only when the cost of staying silent is higher than the cost of the interruption. Both technical and human observations are held to this standard.

**Flint initiates. The developer responds if they want to.**
Flint decides when to speak — not the developer. But when it does speak, a brief response window opens. The developer can do nothing (card fades), dismiss it, act on it, or ask for one follow-up. Then the window closes. This is not a chat. It is a senior dev making a point and being available for one question before going back to their desk.

What stays silent after Flint speaks: win cards, human intelligence observations, the quiet state. These are never followed up — they are acknowledgments or sensitive signals, not invitations.

What opens a brief follow-up window: technical observations, vibe coding acceptance risk, intent questions. One response from Flint, scoped to the observation. Then closed.

**Privacy before everything.**
All data — technical and human — stays local. Nothing leaves the machine without explicit consent. The human layer is opt-in separately from technical awareness.

**Same voice for everything.**
Technical observation and human observation sound identical. Direct, brief, senior dev tone. No softer language for human observations — that would feel condescending. The same respect, the same directness.

---

## 4. Awareness & Human Layer Settings

### Technical awareness levels

**Spark** — sees only what you explicitly show it. No daemon. Manual tools only.

**Flame** — daemon watches the current project. File activity, git behaviour, dependencies, patterns.

**Forge** — daemon watches everything. All projects, full history, cross-repo patterns.

### Human intelligence layer (separate opt-in)

Off by default. Enabled independently of technical awareness at any time:

```bash
flint human on     # enable human intelligence layer
flint human off    # disable it
flint human status # see what it is currently tracking
```

When enabled, the same daemon that watches technical signals also watches behavioural and human patterns. Same sidebar. Same voice. Same time-gating rules. The developer chooses when they are ready for this layer.

**What enabling it means:**
Flint starts tracking session patterns, energy signals, career activity distribution, interruption frequency, and cognitive load indicators. All local. All private. All deletable.

**What it never does:**
Makes clinical observations. Diagnoses conditions. Stores anything identifiable beyond behavioural patterns. Shares anything with anyone.

---

## 5. The Two Planes of Intelligence

### Plane 1 — Technical Intelligence

**Code quality**
Function complexity, naming clarity, duplicate logic, error handling gaps, consistency across the codebase, architectural signals.

**Codebase health over time**
Technical debt accrual rate — not just the debt, the rate at which it is growing. Cognitive load accumulation — the codebase getting harder to hold in your head as abstractions leak. Documentation decay — docs that are likely stale based on what changed around them.

**Knowledge distribution**
Knowledge silos forming in real time — parts of the codebase only one person understands. Single points of failure in human knowledge, not just in architecture.

**Security and dependencies**
CVEs, lock file inconsistencies, zero-maintenance dependencies.

**Git and commit behaviour**
Commit quality, frequency, size, vagueness. Pre-commit consequence mapping.

**Test coverage**
New code without tests. Critical paths modified without coverage.

### Plane 2 — Human Intelligence (opt-in)

**Energy and focus patterns**
Session length over time. Commit quality distribution across the day — most developers make measurably worse technical decisions in the afternoon. Flint can surface this: "your commits after 3pm have a higher revert rate — might be worth saving complex work for mornings."

**Burnout early signals**
Not a diagnosis. A pattern. Shorter sessions over several weeks. Increasing revert rate. More time spent on the same file with less net progress. Longer gaps between sessions. Flint surfaces the pattern once, gently, and does not repeat it.

```
Your sessions have been shorter lately and your revert rate
is up. That sometimes means something's getting in the way.
Just worth knowing.
```

**Interruption cost**
Flint can see when a session is broken — a gap mid-session, a return after 20 minutes away, rapid context switching between unrelated files. It can quantify what this costs: "you've been interrupted 4 times today. Recovery from each takes about 15 minutes of sub-optimal output based on your patterns." Surfaces this as information, not criticism.

**Career blind spots**
Flint sees everything the developer touches. Over weeks and months it builds a picture of what they never touch — infra if they're a frontend dev, security if they're a backend dev, tests if they're either. Surfaces this rarely, as a genuine observation: "you haven't touched anything test-related in 6 weeks across three projects. Not a problem — just worth knowing if it's intentional."

**Cognitive load accumulation**
The codebase is getting harder to understand — not because it's bigger but because the abstractions are leaking. Flint tracks this as a trend and surfaces it when the rate of change crosses a threshold. Also tracks it at the developer level — how long it takes to get back into flow on a codebase after time away is a proxy for cognitive load.

**Growth and stagnation signals**
Is the developer working in genuinely new territory or cycling through the same types of problems? Stagnation is not always visible from inside it. Flint can see the pattern across sessions: "you've been solving the same category of problem for a while. Might be worth deliberately stretching into something adjacent."

**The lone wolf signal**
A developer who never commits to shared branches, never touches shared files, never appears to interact with other parts of a codebase. Flint surfaces this once: "most of your work has been isolated from the rest of the codebase lately. Not always a problem — but worth checking if it's intentional."

**The second system signal**
When rewriting something, developers unconsciously over-engineer it — solving the last problem plus imaginary future ones. Flint detects this from complexity and scope signals on rewrites: "this rewrite is getting significantly more complex than what it's replacing. Sometimes that's right — worth checking it's solving the actual problem."

**Impostor syndrome patterns** (Forge, long observation window)
Over-commenting, excessive defensive coding, refactoring things that don't need it, never pushing code without extensive review. These are real behavioural patterns in code. Flint surfaces this extremely rarely, with maximum care: "you tend to add a lot of defensive layers to code that looks solid. Sometimes that's good practice — sometimes it's worth asking if you trust your own work."

---

## 6. The Sidebar — Editor Presence

### Quiet state
A 4px accent strip on the left edge of the editor. Almost invisible. Just enough to know Flint is there.

### Speaking state — the card

A card slides in from the side — anchored to the editor edge, never a popup or modal. Three sections:

**The observation** — one to three sentences, senior dev tone. Plain language. Direct. No icons, no warning levels, no colour coding.

**The agent prompt** (technical observations only) — a ready-to-use prompt scoped to the exact issue, formatted for the developer's connected agent. Two actions: Copy and Send.

**The follow-up window** (technical observations and vibe coding only) — a single "tell me more" option. Tapping it sends the observation context to Flint and streams one focused follow-up response in the same card. The window then closes permanently for this observation. Not a chat. One question, one answer, done.

Card fades after 8 seconds, after any action, or on next keypress.

### Card types and their behaviours

| Card type | Agent prompt | ✕ Dismiss | Follow-up | Fades |
|---|---|---|---|---|
| Technical observation | ✓ | ✓ (calibrates) | ✓ (one response) | 8s / keypress |
| Vibe coding observation | ✓ | ✓ (calibrates) | ✓ (one response) | 8s / keypress |
| Intent question | — | — | ✓ (one response) | 8s / keypress |
| Win acknowledgment | — | — | — | 5s / keypress |
| Human intelligence | — | — | — | 8s / keypress |
| Taste observation | — | ✓ | ✓ (one response) | 8s / keypress |

### The follow-up — how it works

The developer taps "tell me more." The card expands. Flint streams a focused follow-up — more detail, more context, a concrete example, whatever the observation calls for. Maximum 150 words. Then the card shows only a close button. No further follow-up is possible from the same observation.

```
┌─────────────────────────────────────────────┐
│                                             │
│  That function on line 47 will throw if     │
│  `user` comes back null — happens more      │
│  than you'd think on first login.           │
│                                 [ ✕ ]       │
│  ─────────────────────────────────────────  │
│                                             │
│  In auth.js line 47, add a null check...    │
│                                             │
│  [ Copy ]  [ Send to Cursor ]               │
│  [ Tell me more ]                           │
│                                             │
└─────────────────────────────────────────────┘

                  ↓ after "tell me more"

┌─────────────────────────────────────────────┐
│                                             │
│  That function on line 47 will throw if     │
│  `user` comes back null — happens more      │
│  than you'd think on first login.           │
│                                             │
│  ─────────────────────────────────────────  │
│                                             │
│  The specific risk: `.getId()` is called    │
│  immediately after the DB lookup with no    │
│  guard. On first login, `findUser()` can    │
│  return null when the session hasn't been   │
│  initialised yet — typically within the     │
│  first 200ms of a new session. Add:         │
│  `if (!user) return redirect('/login')`     │
│  before the `.getId()` call.                │
│                                             │
│  [ Close ]                                  │
│                                             │
└─────────────────────────────────────────────┘
```

### What the sidebar never has
- A persistent input box
- Unbounded chat history
- Multiple follow-up rounds from a single observation
- Notification count or badge
- Softer language for human observations
- More than three actions at once
- Anything that turns Flint into a conversation partner

### Agent selection
Set during `flint init`: Cursor, Antigravity, Windsurf, GitHub Copilot, Custom, or None.

---

## 7. The Daemon — Technical Intelligence

### Stillness first
The daemon never reacts to code being actively written. Stillness = no keystrokes for 90 seconds or file switched away from.

### Time-gating rules
- **Minimum gap between observations: 20 minutes**
- **Maximum observations per session: 5**
- **Cooldown per category: 24 hours**
- **First 15 minutes always silent**
- **Rapid file switching = exploration mode — silent**

### Session types — how Flint reads the room

Flint detects four session types. The type determines which evaluation path runs.

**Human coding** — developer is writing the code. Standard daemon behaviour, 90s stillness rule.

**Agent-assisted** — human writes direction, agent helps with specific pieces. Mix of human keystrokes and occasional large agent chunks.

**Vibe coding** — agent writes most of the code, developer directs and accepts. Large file changes without preceding keystrokes, multiple files changing simultaneously. Flint switches to agent output evaluation mode — evaluates completed chunks rather than gradual changes. Stillness redefined as 10 seconds of quiet after a large write. Adjusted time gates (10-minute gap, 8 observations max per session).

**Mixed** — session transitioned between types. Common — developer starts typing, enables agent, shifts to vibe coding mid-session. Flint detects the shift silently.

### Vibe coding mode observations
- **Error handling gaps** in agent-written async calls, API calls, DB queries
- **Edge cases unhandled** — null, empty, zero, negative
- **Missing test coverage** on agent-written logic
- **Security patterns missed** — injection, missing auth, exposed values
- **Codebase inconsistency** — agent output contradicts surrounding conventions
- **Over-engineering** — complexity far beyond what the surrounding code suggests
- **Acceptance risk** — large change accepted in under 30 seconds → surfaces the top 3 things to check
- **Agent prompt quality** — vague prompt detected → one observation, never repeated

### Session states (within any type)
**Exploration** — rapid switching → threshold raised, near-silent
**Deep work** — focused on one or two files → normal threshold
**Debugging spiral** — repeated edits and reverts → near-silent, exception for known errors
**Finishing** — small cleanup changes → threshold raised
**Early session** — first 15 minutes → always silent

### Observation types (all session types)
**Code quality** — complexity, naming, duplicate logic, error handling, consistency
**Git behaviour** — commit frequency, size, message quality, uncommitted work
**Dependencies** — CVEs (always immediate), lock file, maintenance signals
**Test coverage** — new code without tests, critical paths unprotected
**Codebase health** — debt accrual rate, cognitive load accumulation, doc decay, knowledge silos
**Asking questions** — ambiguous intent, max one per session
**Celebrating wins** — genuine improvements, max once per session
**Taste** — readability and naming, low frequency, always hedged
**Deliberate silence** — deep concentration, Flint steps back

Full observation detail and vibe coding skeletons in MVP.md and ARCHITECTURE.md.

---

## 8. The Daemon — Human Intelligence (opt-in)

All human intelligence observations use the same time-gating rules as technical. Additionally:

- **Human observations: maximum 1 per session** — regardless of what is detected
- **Human observation cooldown: 7 days per category** — the same human pattern does not surface twice in a week
- **Human observations never repeat if ignored** — if the developer does not act on it, Flint does not bring it up again in the same form

### What it tracks
- Session length and frequency trends
- Commit quality distribution across time of day
- Revert rate trends over weeks
- Interruption frequency and recovery patterns
- Career activity distribution (what is never touched)
- Cognitive load indicators (time to flow, complexity trends)
- Behavioural patterns in code (over-commenting, defensive coding)
- Growth vs stagnation signals (problem type variety over time)
- Collaboration signals (isolation vs codebase interaction)

### What it never tracks
- Keystrokes, mouse movements, or anything outside coding activity
- Personal information of any kind
- Anything not derivable from file changes, git activity, and session timing

### Storage
All human intelligence data stored in `~/.flint/human.db`. Separate from technical observations. Deletable independently with `flint human --clear`.

---

## 9. Trust & Being Wrong

### Flint never blocks — only advises
Every observation is advisory. Every checkpoint is a pause. The developer always decides.

### Confidence in the voice
Internal confidence threshold: 0.75 minimum to speak. Human observations require 0.85 — higher bar because the stakes are more personal.

High confidence states plainly. Lower confidence hedges naturally. Human observations always use measured, observational language — never clinical, never certain.

### One-tap dismissal (technical observations)
Small ✕ on every technical card. Logs locally. Recalibrates threshold for that category.

### Personal calibration
Stored in `~/.flint/calibration.db`. Per category, per developer. After 3 dismissals in a category within 7 days — threshold rises by 0.05. Slow and conservative by design.

### What Flint never does
- Blocks any action
- Repeats a dismissed observation in the same session
- Says "I told you so"
- Overstates confidence
- Makes clinical observations about the developer
- Penalises disagreement
- Shares anything with anyone

---

## 10. Target Users

| Persona | Technical pain | Human pain |
|---|---|---|
| Junior dev | Doesn't know what they don't know | Impostor syndrome patterns, no growth signal |
| Mid-level dev | Drowns in work around the code | Burnout approaching, career blind spots forming |
| DevOps engineer | Config mistakes invisible until disaster | Lone wolf patterns, high interruption cost |
| Security engineer | Auditing is manual and inconsistent | Stagnation in the same problem type |
| Data scientist | Code quality is an afterthought | Context loss on long-running projects |
| Solo founder / indie dev | No team to catch mistakes | No human signal at all — Flint is the only presence |
| Non-technical PM / founder | Can't understand engineering output | — |

**Primary target for v1:** Mid-level developers (2–5 years) working solo or in small teams. Specifically — developers good enough to ship, but without a senior dev consistently looking over their work — technically or humanly.

---

## 11. Platform Architecture

### Technology decision — Go + TypeScript monorepo

Flint uses two languages with a shared contract, not one language doing everything.

**Go** — daemon, CLI, all background processing. Single compiled binary per platform. Sub-100ms startup. Under 30MB idle RAM. Cross-compiles to all platforms in one command.

**TypeScript** — VS Code/Cursor/Windsurf extension. The only real option for the VS Code API. One build, three marketplaces.

**npm wrapper** — thin Node.js shim that detects OS/arch on install, downloads the right Go binary from GitHub Releases, and exposes `flint` as a standard npm global. Same pattern as esbuild, Turbo, Biome.

**Shared schema** — `shared/schema.json` defines every IPC message between Go and TypeScript. Types are generated for both sides. If the schema changes, both sides break at compile time.

---

### CLI and extension are independent — better together

Neither the CLI nor the extension requires the other to function. They are designed to work alone and to improve each other when combined.

**CLI alone**
Full manual toolkit. Pre-commit hooks. Stakeholder translation. Error dictionary. All terminal-native workflows. The daemon runs and logs observations locally — but without the extension connected, observations are stored silently rather than surfaced in the editor. Everything you invoke explicitly works perfectly.

**Extension alone (standalone mode)**
Manual tools panel at Spark level — direct Anthropic API calls from the extension, no daemon needed. The developer opens the panel, picks a tool, pastes input, gets output. No passive intelligence, no pre-commit checkpoints, no Flame/Forge features. The status bar shows `Flint (standalone)`.

**Both together**
Complete experience. Passive observations surface in the sidebar. Session state awareness. All card types. Full Flame/Forge capability. Daemon observations broadcast simultaneously to terminal and editor.

```
CLI alone        →  full manual + pre-commit + terminal workflows
Extension alone  →  manual tools panel (Spark level only)
Both together    →  complete Flint — everything works
```

---

### Multi-user and multi-workspace resolution

Multiple developers, multiple workspaces, multiple Flint versions — the extension resolves which CLI to use in a deterministic order:

```
1. Workspace-local  →  ./node_modules/.bin/flint
2. Project config   →  .flintrc { "cliPath": "/custom/path" }
3. Global npm       →  $(npm root -g)/bin/flint
4. PATH             →  which flint
5. Not found        →  standalone mode, nudge to install
```

The extension checks this on activation and on every workspace change. The status bar always shows which Flint is active and its version — `Flint 1.2.3 (global)` or `Flint 1.1.0 (local)`.

**Per-workspace daemon isolation:**
Each workspace gets its own daemon socket — `~/.flint/daemon-{workspace-hash}.sock`. Two VS Code windows with different projects run separate daemons and never share state or step on each other.

**Per-workspace config:**
Each project can have a `.flintrc` at its root that overrides the global `~/.flint/config.json` for that project — different role, different awareness level, different agent. Team projects can commit a `.flintrc` so all developers on the project get the same Flint configuration automatically.

---

### Contextual cross-promotion

Cross-promotion is tied to relevant moments — never a timer, never random, never more than once per trigger type.

**CLI → Extension nudge (developer is getting value from CLI, extension would make it better):**
- Developer runs any manual tool 3+ times in a single session → "you'd get this automatically in the editor — install the extension"
- Developer runs `flint check` manually → "the extension runs this on every commit automatically"
- Developer runs `flint debug` on the same error type twice → "the extension would have caught this pattern earlier"

**Extension → CLI nudge (developer tries something that needs the daemon):**
- Extension open for 10+ active minutes with no daemon found → "install the CLI to enable passive intelligence"
- Developer selects a Flame/Forge tool in the panel → "this feature requires the CLI daemon — `npm install -g flint && flint init`"
- Developer opens a new project with no `.flintrc` → "run `flint init` in this project to enable full intelligence"

**Rules for both:**
- Each nudge type fires at most once per day
- Never interrupts active work — appears as a subtle status bar message or a card that fades, never a modal
- Always includes the exact command or link needed, never just "install it"
- Suppressible permanently: `flint config --no-nudge` or clicking "don't show again"

---

### Extension operating modes

**Connected mode** — daemon socket found and alive
Full experience. Passive observations, all card types, Flame/Forge features, session awareness. Status bar: `Flint 1.2.3 (flame) ●`

**Reconnecting mode** — daemon was connected, socket dropped
Extension shows reconnecting state. Retries every 3 seconds. Manual tools still available. Status bar: `Flint (reconnecting...) ○`

**Standalone mode** — no daemon found on activation
Manual tools panel only. Spark level. Contextual nudge to install CLI. Status bar: `Flint (standalone) ○`

**Degraded mode** — daemon found but version mismatch
Extension warns about version mismatch. Manual tools available. Passive features disabled until versions align. Status bar: `Flint (version mismatch) ⚠`

---

### System diagram

```
┌──────────────────────────────────────────────────────────┐
│                   Flint Core Engine (Go)                  │
│    Technical + Human intelligence · SQLite (~/.flint/)    │
└──────┬───────────────────────────────────────────────────┘
       │  ~/.flint/daemon-{workspace-hash}.sock
       │  (one socket per workspace)
  ┌────┴─────────────────────────────────────────────┐
  │                                                  │
  ▼                                                  ▼
CLI (terminal)                          VS Code / Cursor / Windsurf
manual tools                            extension (TypeScript)
pre-commit hook                         ├── connected mode (full)
`flint check`                           ├── standalone mode (Spark only)
`flint update`                          ├── reconnecting mode
error dictionary                        └── degraded mode (version mismatch)
                                        status bar: version + awareness + state

  └─────────────────────┬────────────────────────────┘
                        │
               ┌────────▼────────┐
               │  ~/.flint/      │
               │  SQLite store   │
               │  (shared)       │
               └─────────────────┘
                        │
          ┌─────────────▼──────────────┐
          │   Notification layer (v2+) │
          │  Slack · Email · Telegram  │
          └────────────────────────────┘
```

See `ARCHITECTURE.md` for full system architecture, all skeletons, and version-by-version evolution.

---

## 12. Editor Support Roadmap

**v1** — CLI + VS Code + Cursor + Windsurf (one build, three marketplaces)
**v2** — JetBrains + Neovim + Antigravity (wait for 2.0 stability)
**v3** — Zed + others

---

## 13. Notification & Stakeholder Channels

**v2** — Slack, Email, Telegram
**v3** — WhatsApp, Discord, Notion/Linear

All channels are one-way broadcasts. Flint speaks. Stakeholders read.

---

## 14. Privacy by Design

- All data local — `~/.flint/` — never on Flint servers
- Nothing sent to API without user-triggered action
- Human layer opt-in, off by default
- Human data stored separately, deletable independently
- Awareness level always visible in sidebar and terminal
- `flint memory` shows everything stored
- `flint memory --clear` deletes everything instantly
- No telemetry without explicit opt-in

---

## 15. Key Principles

- **Both planes, one presence.** Technical and human intelligence through the same voice, same sidebar, same tone.
- **Observe behaviour, never judge character.** Patterns, not diagnoses. Facts, not feelings.
- **Earn every interruption.** Silence is the default on both planes.
- **Flint initiates, not the developer.** Flint decides when to speak. When it does, a brief follow-up window opens — one response, scoped to the observation, then closes. Not a chat. Not a conversation that runs indefinitely.
- **Privacy before features.** Local first. Human data especially.
- **Same voice for everything.** No softer tone for human observations. Same respect, same directness.
- **CLI is the foundation.** Every other surface is a layer on top.

---

## 16. Version Roadmap

Full feature registry in `FEATURES.md`. Summary per version below.

### v1 — Foundation + Intelligence (22 features)
CLI + VS Code + Cursor + Windsurf. Three awareness levels. Human layer opt-in. 24 role-aware manual tools. **`flint repl`** — interactive codebase session. Full daemon — technical and human intelligence. 12-signal session type classifier (human / agent-assisted / vibe coding / pair programming / mixed). Stack-aware footgun library. Onboarding scan (`flint scan`). Explainability on low-confidence observations. Flint Explain (`flint explain`). Flint Diff (`flint diff`). Flint Watch — tripwires. Pre-commit checkpoints. Universal error log. Silent auto-fix engine. Personal error dictionary. Stakeholder translation. Flint-initiates communication model with follow-up window. CLI + extension independence. Contextual cross-promotion. Multi-workspace isolation.

### v2 — Workflow Intelligence + Stakeholder Layer (17 features)
Habit feedback. Daily brief. Tool personalisation. Flint REPL (`flint repl`). `flint pr` — PR review. Dead code detector. Environment parity checker. Dependency health scorecard. Pair programming awareness. Flint Journal. Public observation library (beta). JetBrains + Neovim + Antigravity editors. Slack + Email + Telegram notifications. Web dashboard. REST bridge.

### v3 — Codebase Intelligence + Expanded Reach (14 features)
Codebase indexer (tree-sitter). Deep codebase learning. Cross-project memory. Outcome feedback loop. Real-time doc generation. Systemic security patterns. Architectural opinions. API contract monitoring. Flint Teach. Flint Mood. Agent quality differentiation. Offline / local model fallback. WhatsApp + Discord + Notion + Linear. Zed editor.

### v4 — Team Intelligence (12 features)
Team pattern aggregation. Knowledge distribution alerts. Git blame intelligence. Shared runbooks. Team stakeholder reports. Encrypted P2P sync. Team human intelligence. Async code review layer. Public observation library (full). Agent fingerprint community contributions. License and seat management.

---

## 17. Success Metrics

| Version | Key metric | Target |
|---|---|---|
| v1 | Unprompted observation rate | ≥ 1 useful observation per session |
| v1 | Onboarding scan completion | > 80% of new installs run `flint scan` |
| v1 | Human layer opt-in rate | > 25% of Forge users within 2 weeks |
| v1 | Awareness upgrade rate | > 40% upgrade Spark → Flame within 2 weeks |
| v1 | Session type accuracy | > 85% correct classification on confirmed sessions |
| v1 | Qualitative signal | "It said something I didn't ask for and it was right" |
| v2 | REPL adoption | > 20% of CLI users try `flint repl` within 2 weeks of v2 |
| v2 | Notification activation | > 30% of Forge users connect ≥ 1 channel |
| v3 | Outcome correlation | First observation correctly correlated to future bug |
| v4 | Team accounts | 100 team accounts within 6 months of v4 launch |

---

## 18. Open Questions

- Monetisation — free at Spark, paid at Flame/Forge? Human layer as premium?
- API key model — user brings own key (v1) vs Flint proxies (v2+)?
- How do we communicate the human layer without making developers uncomfortable?
- Daemon resource usage — can we stay under 30MB with 12-signal classifier running?
- Role auto-detection from `package.json`, `Dockerfile`, `requirements.txt`?
- WhatsApp Business API — requires Meta approval, worth starting now?
- What is the right observation window for human intelligence? Weeks? Months?
- How do we handle the impostor syndrome observation — most sensitive one in the product?
- Footgun library contribution process — how do we maintain quality as community grows?
- Agent fingerprinting — do we publish fingerprints publicly or keep them internal?
- Clipboard monitoring — do we even include this given privacy sensitivity?

---
