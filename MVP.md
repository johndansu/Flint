# MVP Scope
## Flint v1 — Foundation + Daemon + Sidebar Presence

**Version:** 0.6
**Status:** Draft
**Goal:** Ship the first dev tool that notices things on its own and sounds like a senior dev

---

## The One Question v1 Must Answer

> "Did Flint say something you didn't ask for — and was it right?"

That's the moment. Not "did it answer my question well." Did it notice something on its own, surface it in plain language, and was it actually useful. Everything in v1 is built toward that moment.

---

## The Two Things That Ship in v1

### 1. The Daemon
A background process that runs continuously. Watches file activity, git behaviour, dependencies, and error patterns. Decides on its own when something is worth saying. No command needed. No trigger needed.

This is what makes Flint different from everything else.

### 2. The Sidebar Presence
How the daemon communicates inside the editor. A slim, almost invisible strip. When the daemon has something worth saying — it appears, states it in plain senior-dev language, then fades. No input box. No reply. One-way.

Together these two things are Flint's identity. The manual tools (review, debug, doc, audit) are still there — but they're the floor, not the ceiling. The daemon and sidebar are the ceiling.

---

## Surfaces in v1

| Surface | Status | Notes |
|---|---|---|
| CLI | v1 — primary | Foundation, daemon runs here |
| VS Code sidebar | v1 | Sidebar presence + daemon integration |
| Cursor sidebar | v1 | Same extension, different marketplace |
| Windsurf sidebar | v1 | Same extension, different marketplace |
| JetBrains plugin | v2 | Separate plugin system |
| Neovim plugin | v2 | Terminal-native, high alignment |
| Antigravity | v2 | Wait for 2.0 to stabilise |
| Zed | v3 | Watch and wait |
| Web dashboard | v2 | Stakeholders only |

---

## CLI and Extension — Independent by Design

Neither requires the other. Both work alone. Together they are complete.

### CLI alone
Full manual toolkit. Pre-commit hook. Stakeholder translation. Error dictionary. All terminal workflows. Daemon runs and logs observations locally — without the extension connected, observations are stored silently rather than surfaced in the editor. Everything you explicitly invoke works perfectly.

### Extension alone — standalone mode
Manual tools panel at Spark level. Direct Anthropic API calls from the extension — no daemon needed. Developer opens panel, picks tool, pastes input, gets output. No passive intelligence. No pre-commit checkpoints. No Flame/Forge features.

Status bar shows: `Flint (standalone) ○`

### Both together
Complete Flint. Passive observations surface in the sidebar. Full session awareness. All card types. Flame/Forge features. Daemon broadcasts simultaneously to terminal and editor.

```
CLI alone        →  manual + pre-commit + terminal workflows
Extension alone  →  manual tools panel (Spark only)
Both together    →  complete Flint
```

---

## Extension Operating Modes — v1 Spec

**Connected** — daemon socket found and alive
Full experience. All features. Status bar: `Flint 1.0.0 (flame) ●`

**Reconnecting** — daemon was connected, socket dropped
Manual tools still available. Retries every 3 seconds. Status bar: `Flint (reconnecting...) ○`

**Standalone** — no daemon found on activation
Spark-level manual tools only. Contextual nudge to install CLI. Status bar: `Flint (standalone) ○`

**Degraded** — daemon found but version mismatch
Manual tools available. Passive features disabled. Status bar: `Flint (version mismatch) ⚠`

---

## Multi-User and Multi-Workspace Resolution — v1 Spec

The extension resolves which CLI to use in this exact order:

```
1. ./node_modules/.bin/flint       ← workspace-local install
2. .flintrc { "cliPath": "..." }   ← project-level override
3. $(npm root -g)/bin/flint        ← global npm install
4. which flint                     ← PATH resolution
5. not found                       ← standalone mode + nudge
```

Checked on activation and on every workspace change. Never cached between workspace switches.

**Per-workspace daemon isolation:**
Each workspace gets its own socket — `~/.flint/daemon-{workspace-hash}.sock`. Two VS Code windows with different projects run completely separate daemons. No shared state. No interference.

**Per-workspace config:**
Each project can have a `.flintrc` at its root that overrides `~/.flint/config.json` for that project:

```json
{
  "role": "devops",
  "awareness": "flame",
  "agent": "cursor",
  "cliPath": "/custom/path/to/flint"
}
```

Teams can commit `.flintrc` to version control. All developers on the project get the same Flint configuration automatically on clone.

---

## Contextual Cross-Promotion — v1 Spec

Never a timer. Never random. Always tied to a moment where the missing piece would have genuinely helped. Each nudge type fires at most once per day. Always suppressible.

### CLI → Extension nudge

**Trigger 1 — repeated manual tool usage**
Developer runs any tool 3+ times in one session.
```
  💡 You'd get this automatically in your editor.
     Install the Flint extension: marketplace.visualstudio.com/...
     (flint config --no-nudge to suppress)
```

**Trigger 2 — manual pre-commit check**
Developer runs `flint check` manually.
```
  💡 The extension runs this automatically on every commit.
     Install the Flint extension: marketplace.visualstudio.com/...
```

**Trigger 3 — repeated same error type**
Developer runs `flint debug` on the same error category twice.
```
  💡 The extension would have caught this pattern earlier.
     Install the Flint extension: marketplace.visualstudio.com/...
```

### Extension → CLI nudge

**Trigger 1 — active session, no daemon**
Extension open for 10+ active minutes with no daemon found.
```
  Flint is in standalone mode. Install the CLI for passive intelligence:
  npm install -g flint && flint init
```

**Trigger 2 — Flame/Forge feature selected**
Developer selects a tool that requires the daemon.
```
  This feature requires the Flint CLI daemon.
  npm install -g flint && flint init
```

**Trigger 3 — new project, no .flintrc**
Developer opens a project Flint hasn't seen.
```
  Run `flint init` in this project to enable full intelligence.
```

### Rules for all nudges
- At most once per trigger type per day
- Never a modal — status bar message or subtle card only
- Always includes the exact command or link
- Suppressible permanently: `flint config --no-nudge` or "don't show again"
- Never fires during active typing or when a card is already showing

---

## Awareness Levels

### Spark
No daemon. Manual tools only. Flint sees what you show it.

### Flame
Daemon active on current project. Sidebar presence active. Watches: file saves, git activity, dependency CVEs, function growth, commit patterns.

### Forge
Daemon active across all projects. Everything in Flame plus: cross-project patterns, full error history, habit tracking, personal error dictionary unlocked.

---

## The Daemon — v1 Spec

### What it watches (Flame)
- **File saves** — function length, complexity, same file touched repeatedly without commit
- **Git diffs** — pre-commit consequence mapping, uncommitted work, commit message quality
- **Dependencies** — CVEs in current versions, lock file inconsistencies
- **Patterns** — recurring code shapes that have caused errors before in this project

### What it watches additionally (Forge)
- Cross-project error signatures and habit patterns
- Personal recurring mistakes across codebases
- Error frequency changes on specific files over time

---

### Time-gating rules — the most important part of the daemon

The daemon never reacts to code being actively written by a human. Half-written functions, incomplete refactors, files mid-change — noise, not signal. Flint waits for stillness.

**Stillness = no keystrokes for 90 seconds, or the developer has moved away from the file.**

Exception: vibe coding mode has its own stillness definition — see below.

Additional time gates:
- **Minimum gap between observations: 20 minutes** — Flint never speaks twice within 20 minutes
- **Maximum observations per session: 5** — after 5, the daemon goes silent for the rest of the session
- **Cooldown per category: 24 hours** — same category does not trigger twice in a day
- **First 15 minutes of every session are silent** — Flint lets the developer get into flow first
- **Rapid file switching = exploration mode** — daemon stays silent when files change faster than every 30 seconds

A 4-hour session might produce 2–3 observations, well spaced. That is the target.

---

### Session state awareness — reading the room

The daemon detects four session types and five states. The session type changes everything about how and when Flint speaks.

---

#### Session types — how Flint detects them

**Human coding**
Steady keystrokes, gradual file changes, natural pauses. The developer is writing the code. Standard daemon behaviour applies.

Detection signals: keystroke cadence > 0, file changes are incremental, single file focus, typing pauses between changes.

**Agent-assisted**
Human writes the direction, agent helps with specific pieces. Mix of human keystrokes and occasional large agent-written chunks. The developer is still in control but accepting agent suggestions.

Detection signals: occasional large file changes (50+ lines) interspersed with human typing, changes come in discrete bursts followed by human review activity.

**Vibe coding**
The agent is writing most of the code. Developer is directing, reviewing, and accepting — not typing. Large chunks of code appear rapidly across multiple files. No gradual growth — files go from empty to complete.

Detection signals: large file changes (100+ lines) without preceding keystrokes, multiple files changing simultaneously, changes arrive faster than any human could type, little to no keystroke activity between file changes.

**Mixed**
Session started as one type and shifted. Common — developer starts by typing, enables agent, shifts to vibe coding. Flint detects the shift and transitions observation mode mid-session.

---

#### Session states (within any session type)

**Exploration** — reading more than writing, rapid switching
Threshold raised significantly. Almost no observations.

**Deep work** — focused on one or two files, steady progress
Normal threshold. Standard cadence.

**Debugging spiral** — repeated edits and reverts, same error recurring
Near-silence. Exception: known error in dictionary surfaces once.

**Finishing** — small changes, cleanup, final saves
Threshold raised. Pre-commit checkpoint fires normally.

**Early session** — first 15 minutes
Always silent. No exceptions.

---

### Vibe coding mode — full spec

When the daemon detects vibe coding, it switches to a completely different observation model. The human-coding rules no longer apply.

**What changes in vibe coding mode:**

**Stillness redefined**
Not 90 seconds of no keystrokes. An agent chunk is complete when file activity stops for 10 seconds after a large write. That is the evaluation trigger.

**Observation target shifts**
In human coding, Flint watches gradual change. In vibe coding, Flint evaluates completed agent outputs — the whole chunk, not the delta.

**Time gates adjusted**
- Minimum gap between observations: 10 minutes (shorter — agent output arrives faster)
- Maximum observations per session: 8 (higher — more agent output to evaluate)
- First 5 minutes silent (shorter than human mode — agent sessions ramp up faster)

**What Flint watches for in vibe coding mode:**

*Things agents commonly miss:*
- Error handling absent on external calls, async operations, DB queries
- Edge cases unhandled — null, empty array, zero, negative numbers
- No test coverage on agent-written logic
- Security patterns missed — raw interpolation, missing auth checks, exposed secrets
- Hardcoded values that should be environment variables

*Things agents get structurally wrong:*
- Solved the stated problem but not the actual problem — detectable when agent output contradicts the surrounding codebase's patterns
- Over-engineered — complexity far beyond what the surrounding code suggests is needed
- Inconsistent with codebase conventions — naming, structure, error handling style differs from the rest of the project
- Dependencies added that duplicate existing ones in the project

*The acceptance risk — the most important vibe coding observation:*
When Flint detects a large agent output accepted quickly (< 30 seconds of review activity after a 100+ line write), it surfaces once:

```
You just accepted a significant change quickly — here are the
three things most worth checking before moving on:
the error handling on the API call, the null case on line 34,
and whether this matches how auth is handled elsewhere.
```

This is the highest-value observation in vibe coding mode. The moment of accepting without fully understanding is where most vibe coding bugs enter.

*Agent prompt quality feedback (vibe coding only):*
Flint observes the prompts being sent to agents (from clipboard activity or connected agent integration). When a prompt is detectably vague:

```
That prompt is quite broad — the agent will make a lot of
assumptions. A more scoped instruction usually gives cleaner output.
```

Maximum once per session. Never repeated.

**What Flint celebrates in vibe coding mode:**
- Developer reviews and rejects agent output — shows critical thinking
- Developer adds tests to agent-written code — excellent signal
- Developer catches an agent inconsistency before Flint does — rare but worth acknowledging

```
Good catch — that wasn't consistent with how the rest of the
codebase handles this. Worth keeping that habit.
```

**Agent prompt generation in vibe coding mode:**
Prompts generated by Flint in vibe coding mode are scoped differently — they target the agent, not the developer. They are review instructions, not fix instructions:

```
Review the function added in auth.js in the last commit.
Check specifically: error handling on the db.findUser call,
the null case when user is not found, and whether the JWT
signing matches the pattern used in the existing /login route.
Fix any issues found. Don't change anything else.
```

---

### Human coding session states (unchanged)

**Exploration** — rapid file switching, reading more than writing
Daemon raises threshold significantly. Developer is thinking, not building. Almost no observations.

**Deep work** — long focus on one or two files, steady progress
Normal threshold. This is the primary coding state. Observations fire at standard cadence.

**Debugging spiral** — repeated edits and reverts on the same lines, frequent saves with no net change, same error appearing multiple times
Daemon raises threshold to near-silence. Developer needs space to think. Exception: if Flint has seen this exact error before in the error dictionary, it surfaces that once and then goes silent.

**Finishing mode** — small changes, cleanup, comment additions, final saves before commit
Daemon raises threshold. Pre-commit checkpoint fires normally.

**Early session** — first 15 minutes regardless of state
Always silent. No exceptions.

The daemon never announces what state or session type it has detected. It adjusts silently.



### Asking questions about intent

Not every observation is a statement. When the daemon sees something that could be intentional or could be a mistake, it asks rather than assumes.

Framed as a question, not a warning:
```
Is the error here intentional — are you handling it upstream,
or is this a path that still needs covering?
```

```
This function mutates the input directly — was that deliberate,
or would you rather it return a new value?
```

Questions are lower frequency than observations. Maximum one question per session. If the developer ignores it, it is never repeated. A question in the sidebar still has no input box — it is a prompt for the developer to think, not a conversation starter.

---

### Celebrating wins

Flint is not only a problem-finder. Pure problem-finding without acknowledgment of good work feels like surveillance. A senior dev occasionally says "that's clean" — Flint does too.

Win signals the daemon recognises:
- A large complex function successfully split into smaller focused ones
- Test coverage added to a previously untested critical path
- A recurring error type that has not appeared in several sessions
- A dependency vulnerability fixed promptly after Flint flagged it
- A commit message that is genuinely descriptive after a period of vague ones
- A long-standing piece of duplicated logic finally consolidated

What a win card looks like:
```
That refactor cleaned things up considerably — the two functions
are much easier to reason about separately.
```

```
Good call adding tests to the auth path. That one gets people.
```

Win acknowledgments are rare — maximum once per session, only when the signal is clear. They never feel forced or automated. They are not congratulations for normal work — only for genuine improvements.

Win cards have no agent prompt and no ✕ dismissal. They are not actionable. They just appear and fade.

---

### Deliberate silence — knowing when to let you work

Flint's silence is not always "nothing to flag." Sometimes it is a considered decision to let the developer figure something out themselves.

When the daemon detects a complex problem-solving session — deep focus, exploratory changes, multiple approaches being tried — it raises its threshold significantly even if it has observations queued. The developer needs space. Flint stepping in would break something more valuable than what it would catch.

This is different from the debugging spiral state. The spiral is frustration — Flint recognises that and stays quiet to avoid adding noise. The complex problem-solving state is concentration — Flint recognises that and stays quiet out of respect.

The daemon never announces this decision. It simply waits.

---

### Taste — aesthetic observations

Flint has occasional opinions about code quality that go beyond correctness. Not style enforcement — that is what linters are for. Genuine aesthetic observations about readability and long-term maintainability.

These are always framed as opinions, never rules. Low frequency — at most once every few sessions per category. Never triggered on code in progress.

Examples:
```
This works, but the name `data` is going to confuse whoever
reads this next — including you in three months.
```

```
The logic here is correct but it's doing a lot in one place.
Not a bug — just worth keeping an eye on as it grows.
```

Taste observations have the lowest confidence threshold display — they always use hedged language. They are never included in the pre-commit checkpoint. They are strictly optional observations about quality, never about correctness.

---

### Memory of why code was written (partial in v1, full in v2)

In v1, Flint reads what it can: git commit messages, inline comments, PR descriptions if available. Before making an observation about a piece of code, the daemon checks whether there is documented context that explains the decision.

If a comment says `// intentional — see issue #47` Flint does not flag that code.
If a commit message explains a constraint, Flint factors it in.

Full "why" memory — understanding undocumented decisions from patterns and history — requires the deeper codebase learning in v2.

---

### Cross-project mistake memory — pushing back (Forge)

When the daemon detects a pattern that caused a problem in a previous project, it does not just flag it as a general observation. It connects it explicitly:

```
You handled async errors this way in payments-service and it
caused silent failures in production. Same pattern here —
worth a closer look before this ships.
```

This requires Forge awareness and the personal error dictionary being populated over time. It gets more powerful the longer Flint has been running.

---

### What the daemon observes (Flame)

**Code quality** — after stillness only
- Function beyond ~50 lines with multiple distinct responsibilities
- Deeply nested conditionals (3+ levels)
- Missing error handling in async calls, external APIs, DB queries
- Genuinely ambiguous names — not style, actual clarity risk
- Duplicate logic in two or more places in the same file

**Git and commit behaviour**
- File edited across 5+ sessions without being committed
- Staged diff touching more than 10 files
- Vague commit messages — Flint suggests better
- Active feature work going days without a commit

**Dependencies** — exception to all time-gating, always immediate
- Known CVE in a current dependency version
- Dependency added but not in lock file
- Zero-maintenance dependency added to a greenfield project

**Test coverage**
- New function or module with no corresponding test file
- Critical path (auth, payments, data deletion) modified without nearby coverage
- Test file untouched while source file changed significantly

**Consistency**
- Same operation handled differently in two parts of the codebase
- Environment variables referenced directly rather than through a config layer
- Hardcoded values appearing more than once that should be constants

---

### What the daemon observes additionally (Forge)

**Cross-session patterns**
- Same error type appearing repeatedly — knowledge gap, not just a bug
- Heavily-edited file across many sessions — highest-risk file in the project
- Pattern made before in a different project appearing here — explicit cross-project callout

**Habit signals** — long cooldowns
- Sessions consistently ending without committing — once per week maximum
- Tool category never used — gentle nudge once, long cooldown
- Consistently brief commit messages — one observation, long cooldown

---

### When the daemon stays silent
- Code actively being written (within 90 seconds of last keystroke)
- First 15 minutes of a session
- Within 20 minutes of the last observation
- 5 observations already made this session
- Same category already triggered today
- Exploration, debugging spiral, or complex problem-solving mode detected
- Documented intent found that explains the code in question
- Low confidence — if unsure, say nothing
- Normal, healthy, clean activity

Silence is the default. Every observation, question, win, and taste note is earned.

---

### Daemon process
- Runs as a lightweight background process started by `flint start`
- Auto-starts on `flint init` if awareness is Flame or Forge
- Resource target: < 1% CPU, < 50MB RAM at rest
- Communicates with editor extension via local socket
- Logs all observations to `~/.flint/observations.db` (local only)
- Session state stored in memory, time-gate state persists across restarts

---

### The Sidebar — v1 Spec

### Quiet state
A 4px accent strip on the left edge of the editor. Almost invisible. Just enough presence to know Flint is there.

### Speaking state — the card

A clean card slides in from the side — anchored to the editor edge, never a popup or modal. Up to three sections depending on card type.

**Section 1 — The observation**
One to three sentences. Senior dev voice. No icons, no warning levels, no colour coding.

**Section 2 — The agent prompt** (technical + vibe coding observations only)
Ready-to-use prompt scoped to the exact issue. Formatted for the developer's connected agent.
Actions: **Copy** | **Send to [Agent]**

**Section 3 — The follow-up window** (technical, vibe coding, intent questions, taste only)
A single **"Tell me more"** option. Tapping it streams one focused follow-up response — more detail, concrete example, specific explanation. Maximum 150 words. Card then shows only **Close**. No further follow-up from the same observation. Ever.

Card fades after 8 seconds, after any action, or on next keypress.

### Card types

| Card type | Agent prompt | ✕ Dismiss | Tell me more | Fades |
|---|---|---|---|---|
| Technical observation | ✓ | ✓ | ✓ | 8s |
| Vibe coding observation | ✓ | ✓ | ✓ | 8s |
| Intent question | — | — | ✓ | 8s |
| Win acknowledgment | — | — | — | 5s |
| Human intelligence | — | — | — | 8s |
| Taste observation | — | ✓ | ✓ | 8s |

### Full card flow

```
┌─────────────────────────────────────────────┐
│  That function on line 47 will throw if     │
│  `user` comes back null — happens more      │
│  than you'd think on first login.  [ ✕ ]   │
│  ─────────────────────────────────────────  │
│  In auth.js line 47, add a null check for   │
│  the `user` object before calling           │
│  `.getId()`. Handle the null case by        │
│  redirecting to /login. Don't change        │
│  anything else in this function.            │
│  [ Copy ]  [ Send to Cursor ]               │
│  [ Tell me more ]                           │
└─────────────────────────────────────────────┘

              ↓ developer taps "Tell me more"

┌─────────────────────────────────────────────┐
│  That function on line 47 will throw if     │
│  `user` comes back null — happens more      │
│  than you'd think on first login.           │
│  ─────────────────────────────────────────  │
│  The specific risk: `.getId()` is called    │
│  immediately after the DB lookup with no    │
│  guard. On first login, `findUser()` can    │
│  return null when the session hasn't been   │
│  initialised yet — typically within the     │
│  first 200ms of a new session. Add:         │
│  `if (!user) return redirect('/login')`     │
│  before the `.getId()` call.                │
│  [ Close ]                                  │
└─────────────────────────────────────────────┘
```

### Agent selection
Set during `flint init` or anytime via `flint config`:

```
Which AI agent do you use?

❯ Cursor
  Antigravity
  Windsurf (agent mode)
  GitHub Copilot
  Custom / other
  None — copy only
```

Flint formats prompts per agent. Cursor and Windsurf accept natural language scoped to a file and line. Antigravity accepts task-level instructions with broader context. Custom outputs a generic prompt.

### What the sidebar never has
- A persistent input box
- Unbounded chat history
- Multiple follow-up rounds from a single observation
- Notification count or badge
- More than three actions at once
- Softer language for human observations
- Anything that turns Flint into a conversation partner



---

## Manual Tools (all awareness levels)

24 role-aware tools. The developer invokes these intentionally. The daemon runs underneath automatically.

| Role | Doc | Review | Debug | Security / Perf |
|---|---|---|---|---|
| Web dev | README generator | Code reviewer | Error explainer | Performance audit |
| Full-stack | README generator | Full-stack audit | Error explainer | Security checker |
| DevOps | Runbook writer | Config reviewer | Infra debugger | Security scanner |
| Cybersecurity | Security report | Code security audit | Incident analyser | Threat modeller |
| Data / ML | Model card writer | Notebook reviewer | Pipeline debugger | Query optimiser |
| Mobile | App store copy | Mobile code review | Crash analyser | Performance checker |

Every tool output ends with one sentence explaining the principle. Senior dev tone throughout.

---

## Human Intelligence Layer — v1 Spec

Off by default. Enabled independently of technical awareness:

```bash
flint human on       # enable
flint human off      # disable
flint human status   # see what it is tracking
flint human --clear  # delete all human data
```

When enabled, the same daemon that watches technical signals also watches behavioural and human patterns. Same sidebar. Same voice. Same time-gating rules with stricter limits:

- **Maximum 1 human observation per session** — regardless of signals detected
- **7-day cooldown per human category** — same pattern does not surface twice in a week
- **Never repeated if ignored** — if the developer does not act on it, Flint does not bring it up again in the same form
- **Confidence threshold: 0.85** — higher bar than technical (0.75) because the stakes are more personal
- **Observation window: minimum 2 weeks of data** — Flint does not make human observations until it has enough history to be confident

### What it tracks

**Energy and focus patterns**
Session length trends over time. Commit quality distribution across the day — if the developer's revert rate is consistently higher in the afternoon, Flint surfaces this once as useful information, not criticism.

```
Your commits after 3pm have a higher revert rate than morning ones.
Might be worth saving the complex decisions for earlier in the day.
```

**Burnout early signals**
Sessions getting shorter over several weeks. Revert rate increasing. More time on the same file with less net progress. Longer gaps between sessions. Flint surfaces this once, gently.

```
Your sessions have been shorter lately and your revert rate is up.
That sometimes means something's getting in the way. Just worth knowing.
```

**Interruption cost**
Mid-session gaps, return after 20+ minutes, rapid context switching between unrelated files. Flint quantifies what it observes — not as a judgment, as data.

```
You've had 4 interruptions today. Your commit quality tends to drop
for about 15 minutes after each one — just something to be aware of.
```

**Career blind spots**
Flint sees everything the developer touches. Over weeks it builds a picture of what they never touch. Surfaces this rarely, at most once a month per category.

```
You haven't touched anything test-related in 6 weeks across
three projects. Not a problem — just worth knowing if it's intentional.
```

**Cognitive load accumulation**
Tracks how long it takes to get back into flow on a codebase after time away — a proxy for how cognitively loaded it is becoming. Surfaces as a codebase health signal, not a personal one.

```
This codebase seems to be taking longer to get back into after breaks.
That usually means the abstractions are getting harder to hold in mind.
```

**Growth and stagnation signals**
Tracks problem type variety across sessions. If the developer is cycling through the same category of problem for weeks, Flint notices.

```
You've been solving the same category of problem for a while.
Might be worth deliberately stretching into something adjacent.
```

**The lone wolf signal**
Isolation from the rest of the codebase — no shared branches, no touches to shared files. Surfaces once.

```
Most of your work has been isolated from the rest of the codebase lately.
Not always a problem — worth checking if it's intentional.
```

**The second system signal**
On rewrites, complexity and scope signals that suggest over-engineering.

```
This rewrite is getting significantly more complex than what it's replacing.
Sometimes that's right — worth checking it's solving the actual problem.
```

**Impostor syndrome patterns** (Forge only, 4+ weeks of data, highest confidence threshold)
Over-commenting, excessive defensive coding, refactoring things that don't need it. Surfaces at most once, with the most careful framing of anything Flint says.

```
You tend to add a lot of defensive layers to code that looks solid.
Sometimes that's good practice — sometimes it's worth asking if you trust your own work.
```

### What it never tracks
- Keystrokes, mouse movements, or anything outside coding activity
- Personal information of any kind
- Anything not derivable from file changes, git activity, and session timing

### Storage
All human data stored in `~/.flint/human.db`. Separate from technical observations. Deletable independently.

### What human observations look like in the sidebar
Same card. Same voice. Same fade behaviour. No special marking that distinguishes them as "human" vs "technical" — they are all just Flint speaking. No agent prompt on human observations. No ✕ dismissal — human observations are not calibration signals, they are information.

---

## Trust & Being Wrong — v1 Spec

Flint will sometimes be wrong. The design must handle this gracefully or developers will uninstall it.

### Flint never blocks
Every observation is advisory. Every checkpoint is a pause. The developer always decides. No exceptions.

### Confidence in the voice
The daemon assigns an internal confidence score to every observation before surfacing it. Threshold to speak: 0.75 minimum.

Below 0.75 — the daemon stays silent regardless of what it found.

The voice reflects confidence level:

High confidence (CVE, known pattern, clear issue):
```
The version of lodash you're running has a known prototype
pollution vulnerability. Worth updating before this ships.
```

Lower confidence (pattern-based, context-dependent):
```
This might not be an issue depending on how you're calling it,
but that null check on line 47 could be skipped on first login.
Worth a second look.
```

### One-tap dismissal

Every sidebar card has a small ✕ in the corner. One tap, no explanation needed. Logs the dismissal locally and recalibrates.

### Personal calibration

Stored in `~/.flint/calibration.db`. Per category, per developer, fully local.

- Dismiss function length observations repeatedly → threshold for that category rises
- Act on CVE alerts consistently → that category stays unchanged
- Dismiss commit frequency nudges → frequency drops for you

After 3 dismissals in a category within 7 days — Flint raises the threshold for that category by 0.05. After 5 consistent positive signals — threshold can lower by 0.02. Calibration is slow and conservative by design.

### What Flint never does
- Never blocks an action — only advises
- Never repeats a dismissed observation in the same session
- Never says "I told you so"
- Never overstates confidence
- Never penalises the developer for disagreeing

---

## Signature Features

### Pre-commit checkpoint (Flame + Forge)
Git hook fires on every commit. Consequence analysis surfaces in terminal and sidebar simultaneously.

```
› git commit -m "refactor auth flow"

  flint › before you commit

  This touches 4 downstream files. Two have no tests.
  Likely affected: middleware/session.js, routes/user.js

  That string interpolation on line 47 — worth parameterising
  before this goes in.

  utils/format.js looks clean, no downstream impact.

  Commit anyway? (y/n)
```

### Personal error dictionary (Forge)
Every debug run stores error signature locally. Next similar error — Flint checks the dictionary first and surfaces the previous fix.

### Stakeholder translation (Forge)
`flint update` — last N commits in plain English, ready to paste anywhere.

### Teaches while it fixes (all levels)
One sentence at the end of every manual tool output explaining the principle behind the fix.

---

## CLI Commands

```bash
# Setup
flint init                    # first run — awareness, role, API key
flint start                   # start the daemon manually
flint stop                    # stop the daemon
flint status                  # awareness level, daemon state, error count
flint awareness               # change awareness level
flint scan                    # index codebase + establish baselines (run once)

# Manual tools
flint review                  # code review
flint debug                   # error explanation
flint doc                     # documentation
flint audit                   # security / performance
flint explain                 # understand unfamiliar code
flint diff                    # explain a git diff
flint                         # interactive picker

# Signature features
flint check                   # manual pre-commit check
flint update                  # stakeholder translation
flint update --format slack
flint update --format email
flint update --audience cto

# Error log
flint errors                  # recent error log (last 7 days)
flint errors --all            # full history
flint errors --project        # current project only
flint errors --type cve       # filter by error type
flint errors --status unfixed # unresolved errors only
flint errors --clear          # delete error log

# Fix commands
flint fix <error-id>          # fix a specific logged error
flint fix --auto cve          # enable auto-fix for CVE category
flint fix --auto import       # enable auto-fix for import errors
flint fix --auto format       # enable auto-fix for formatting failures
flint fix --auto off          # disable all auto-fix
flint fix --log               # show history of all auto-fixes

# Tripwires
flint watch "pattern"         # set tripwire on pattern
flint watch list              # show all active tripwires
flint watch remove <id>       # remove a tripwire

# Human intelligence layer
flint human on                # enable human intelligence
flint human off               # disable human intelligence
flint human status            # see what it is tracking
flint human --clear           # delete all human data

# Memory + transparency
flint memory                  # everything Flint has stored
flint memory --clear          # delete everything
flint memory --export         # export as JSON

# Config
flint config                  # update role, awareness, API key, agent
flint config --no-nudge       # suppress cross-promotion nudges
```

---

## First Run Experience

```bash
› npm install -g flint
› flint init

  Welcome to Flint.

  Flint is a senior dev that never leaves. It watches your work
  and speaks up when it has something worth saying — without
  being asked.

  How much should Flint see?

  ❯ Spark  — only what you show it (no background watching)
    Flame  — your current project (recommended)
    Forge  — everything (full experience)

  Your role?

  ❯ Web dev · Full-stack · DevOps / cloud
    Cybersecurity · Data / ML · Mobile dev

  Anthropic API key: sk-ant-...

  Which AI agent do you use?

  ❯ Cursor
    Antigravity
    Windsurf (agent mode)
    GitHub Copilot
    Custom / other
    None — copy only

  One more thing — Flint can also observe human patterns in
  how you work: energy levels, burnout signals, career blind
  spots, interruption cost. All local, all private, opt-in only.

  Enable human intelligence layer?

  ❯ Not now — I can enable it later with `flint human on`
    Yes — enable it now

  ✓ Flint is ready.
    Awareness: Flame · Role: Web dev · Agent: Cursor
    Human layer: off · Daemon: running · Editor: VS Code detected

  Flint will speak up when it has something worth saying.
  You don't need to do anything.
```

---

## What Is NOT in v1

| Thing | Reason | Version |
|---|---|---|
| Habit feedback / daily brief | Needs history depth | v2 |
| Slack / Email / Telegram | Validate core first | v2 |
| JetBrains / Neovim | Separate plugin systems | v2 |
| Antigravity | Wait for 2.0 stability | v2 |
| Web dashboard | Devs first | v2 |
| Deep codebase learning | Context strategy needed | v3 |
| Cross-project memory | Needs codebase learning | v3 |
| Real-time doc gen | Complex editor integration | v3 |
| Team features | Solo first | v4 |
| Auto role detection | Not blocking | v1.5 |
| Billing / paywall | Validate value first | v2 |

---

## Tech Stack

### The core decision
Go for runtime. TypeScript for interface. One monorepo with a shared JSON schema contract.

| Layer | Technology | Reason |
|---|---|---|
| CLI | Go + cobra | Single binary, sub-100ms startup, no runtime |
| Daemon | Go + fsnotify | Long-running, low memory, native OS file events |
| IPC | Unix socket (Go stdlib) | Fast local comms, no external dependency |
| VS Code / Cursor / Windsurf | TypeScript + VS Code API | One build, three marketplaces |
| npm distribution | Node.js shim + Go binary | Same pattern as esbuild, Turbo, Biome |
| Local store | SQLite (modernc.org/sqlite) | Pure Go, no CGO, local-first |
| AI | Anthropic API (direct HTTP + streaming) | Full control over streaming |
| Schema codegen | Node.js script | Generates Go + TS types from shared/schema.json |
| Build | Makefile + GoReleaser | Cross-compile all platforms in one command |
| Distribution | npm + GitHub Releases + VS Code Marketplace | Where developers look |

Full rationale and all skeletons in `ARCHITECTURE.md` and `TECH_STACK.md`.

---

## Build Checklist

**Shared core**
- [ ] Tool definitions (role + category + system prompt) as JSON
- [ ] Footgun library (`shared/footguns.json`) — initial entries for all 6 roles
- [ ] Anthropic API wrapper with streaming
- [ ] Awareness level gating (Spark / Flame / Forge)
- [ ] SQLite store — all 9 tables including error_log and auto_fixes
- [ ] Human-readable `~/.flint/errors.log` mirror
- [ ] Senior dev tone layer (applied to all outputs)

**Error log + auto-fix engine**
- [ ] Universal error logger — captures all 7 sources (terminal, test, build, runtime, precommit, dependency, install)
- [ ] Error signature computation (hash of type + stack pattern)
- [ ] Error status tracking (observed / fixed / ignored / auto_fixed)
- [ ] Auto-fix engine with staged-only constraint (never commits)
- [ ] CVE auto-fix (update dependency to patched version)
- [ ] Import auto-fix (unambiguous missing import resolution)
- [ ] Format auto-fix (run project formatter on failure)
- [ ] Auto-fix notification card (always shown, no agent prompt)
- [ ] Auto-fix history log (`flint fix --log`)
- [ ] `flint errors` command with all filters
- [ ] `flint fix` command (manual + auto-fix config)
- [ ] Error dictionary builder (signature → fix mapping from error_log)

**Daemon**
- [ ] Background process scaffold
- [ ] File watcher (chokidar)
- [ ] Git diff monitor
- [ ] Dependency CVE checker
- [ ] Function complexity tracker
- [ ] Commit pattern monitor
- [ ] Session type detector (human / agent-assisted / vibe coding / mixed)
- [ ] Session state detector (exploration / deep work / debugging spiral / finishing / early)
- [ ] Vibe coding mode — agent chunk completion detection (10s stillness after large write)
- [ ] Vibe coding mode — agent output evaluator (error handling, edge cases, security, consistency)
- [ ] Vibe coding mode — acceptance risk detector (large change accepted < 30s)
- [ ] Vibe coding mode — agent prompt quality observer
- [ ] Vibe coding mode — adjusted time gates (10min gap, 8 max per session)
- [ ] Session type transition detection (human → vibe mid-session)
- [ ] Intent reader (commit messages + inline comments before flagging)
- [ ] Win signal detector (refactors, test coverage, CVE fixes)
- [ ] Taste observation engine (readability, naming, structure)
- [ ] Question generator (ambiguous intent detection)
- [ ] Deliberate silence logic (deep concentration detection)
- [ ] "When to speak" decision logic with state-aware thresholds
- [ ] Confidence score system (internal, 0.75 base threshold)
- [ ] Local socket server (communicates with editor extension)
- [ ] Observation logger (`~/.flint/observations.db`)
- [ ] Calibration store (`~/.flint/calibration.db`)
- [ ] Per-category threshold adjustment on dismissal

**Daemon — human intelligence (opt-in)**
- [ ] Session length and frequency tracker
- [ ] Commit quality distribution by time of day
- [ ] Revert rate trend monitor
- [ ] Interruption detection and cost estimator
- [ ] Career activity distribution tracker
- [ ] Cognitive load indicator (time-to-flow proxy)
- [ ] Growth vs stagnation signal (problem type variety)
- [ ] Lone wolf detector
- [ ] Second system signal detector
- [ ] Impostor syndrome pattern detector (Forge, 4+ weeks)
- [ ] Human observation confidence threshold (0.85)
- [ ] 7-day cooldown per human category
- [ ] Human data store (`~/.flint/human.db`)
- [ ] `flint human on/off/status/--clear` commands
- [ ] `flint start` / `flint stop` / `flint status`
- [ ] `flint awareness` / `flint config`
- [ ] Core toolkit commands
- [ ] Interactive picker
- [ ] `flint check` — pre-commit consequence analysis
- [ ] Git pre-commit hook installer
- [ ] `flint update` — stakeholder translation
- [ ] Error dictionary (log + lookup)
- [ ] `flint memory` commands
- [ ] Publish to npm

**Extension — CLI resolution + modes**
- [ ] CLI resolver (`extension/src/cli/resolver.ts`) — 5-step resolution order
- [ ] Per-workspace socket path (`extension/src/ipc/socket.ts`)
- [ ] Mode manager (`extension/src/mode.ts`) — connected / reconnecting / standalone / degraded
- [ ] Status bar — all four mode states with correct icons and tooltips
- [ ] Extension entry point with full mode lifecycle and workspace change listener
- [ ] Cross-promotion engine (`extension/src/nudge/cross-promotion.ts`)
  - [ ] Standalone active session nudge (10+ minutes, once per day)
  - [ ] Flame/Forge feature needs daemon nudge
  - [ ] New project no .flintrc nudge
  - [ ] "Don't show again" suppression persisted in globalState
  - [ ] Daily reset of shown nudges

**CLI — cross-promotion**
- [ ] Nudge engine (`core/internal/nudge/nudge.go`)
- [ ] Tool usage counter in store (per tool, per day)
- [ ] Nudge after 3+ uses of same tool in a day
- [ ] Nudge after manual `flint check`
- [ ] Nudge after repeated same-error debug
- [ ] `flint config --no-nudge` suppression flag
- [ ] All nudges appear after output, never interrupt
- [ ] Extension scaffold (one build covers VS Code, Cursor, Windsurf)
- [ ] IPC client — Unix socket
- [ ] Generated message types from schema
- [ ] Sidebar provider — quiet state (4px strip)
- [ ] Card type router (technical / vibe / intent / win / human / taste)
- [ ] Technical card — observation + agent prompt + Copy + Send + Tell me more + ✕
- [ ] Follow-up window — streams one Anthropic response, shows Close, locks further follow-up
- [ ] Win card — observation only, fades 5s, no actions
- [ ] Human card — observation only, no ✕, no follow-up, fades 8s
- [ ] Intent question card — observation + Tell me more, no ✕
- [ ] Taste card — observation + Tell me more + ✕
- [ ] One-tap ✕ dismissal with IPC feedback to daemon
- [ ] Agent send integration (Cursor, Antigravity, Windsurf, Copilot, Custom)
- [ ] Status bar item (awareness level + session type)
- [ ] Manual tool panel (role selector + tool list + streaming output)
- [ ] Publish to VS Code Marketplace, Cursor registry, Windsurf registry

---

## Timeline

| Week | Milestone |
|---|---|
| 1 | Shared core + schema + all 24 tool prompts + footguns.json + tools.json |
| 2 | CLI — Spark level, core toolkit, `flint init`, `flint scan` (onboarding) |
| 3 | CLI — `flint explain`, `flint diff`, `flint watch`, `flint check`, `flint update` |
| 4 | Daemon — file watcher, git monitor, CVE checker, 12-signal session classifier |
| 5 | Daemon — technical engine, vibe coding mode, acceptance risk, explainability |
| 6 | Daemon — win signals, taste engine, question generator, deliberate silence |
| 7 | Human intelligence layer — all trackers, 0.85 threshold, 7-day cooldown |
| 8 | Extension — CLI resolver, operating modes, sidebar, all card types, follow-up window |
| 9 | Extension — agent integrations, cross-promotion engine, status bar all modes |
| 10 | Daemon ↔ extension socket integration + full end-to-end testing |
| 11 | Forge features — error dictionary, cross-project pushback, personal baseline |
| 12 | Internal testing — daemon useful without annoying? Human layer trusted? Detection accurate? |
| 13 | Soft launch — 20 real developers across human, vibe coding, and pair sessions |
| 14–15 | Gather signal, tune thresholds, tune classifier, plan v2 |

---

## The Hardest Problems in v1

**1. The daemon's "when to speak" threshold**
Too loud → developers disable it on day two. Too quiet → it feels broken. Target: 1–3 observations per session, each worth pausing for. Week one will be wrong. The soft launch feedback loop is what tunes it.

**2. Session type classifier accuracy**
The 12-signal classifier needs to correctly distinguish human from agent from vibe coding from pair programming. Cold start accuracy (first 10 sessions) is lower — observations are held to 0.85 threshold during this period. Pair programming detection is the hardest — 65% cold start accuracy target. Getting this wrong makes every subsequent observation less trustworthy.

**3. The onboarding scan baseline**
`flint scan` needs to establish what "normal" looks like for this specific codebase before the daemon starts making observations. Too aggressive a baseline and early observations are all noise. Too conservative and Flint misses real issues from day one. This is the gate between "interesting tool" and "tool I kept."

**4. Footgun library quality**
The stack-aware footgun library ships with v1. Its quality determines whether role-specific observations feel precise or generic. Every entry needs to be: genuinely surprising (not obvious), correctly scoped to the framework version, and actionable. Bad entries erode trust. This is a content problem more than an engineering problem — worth spending real time on the initial library before shipping.

These four problems, solved well, make v1 indispensable. Solved poorly, they make it another tool developers uninstall after a week.

---

---

## Monorepo Quick Reference

```bash
# First time setup
git clone https://github.com/flint-dev/flint && cd flint
npm install
node scripts/generate-schema.js
make build

# Daily dev
make dev              # daemon (Go watch) + extension (TS watch)
make test             # all Go + TS tests
cd core && go test ./internal/daemon/... -v  # daemon tests only

# Release
git tag v1.0.0 && git push origin v1.0.0
# → CI cross-compiles + publishes npm + publishes extension automatically
```

Full architecture, all skeletons, and version-by-version evolution in `ARCHITECTURE.md`.
