# Session Detection Architecture
## Flint — Multi-Signal Writer Classification

**Version:** 1.0
**Last updated:** May 2026
**Scope:** How Flint determines who or what is writing code

---

## The Problem

Knowing *who* is writing code changes everything about how Flint behaves.

A human in deep work needs silence and occasional observations.
An AI agent producing a large chunk needs a quality review after it completes.
A vibe coding session needs acceptance-risk monitoring.
A pair programming session needs a different kind of awareness entirely.

A single heuristic — "large change without keystrokes = agent" — is too crude. It misses fast human writers, incremental agents, pair programmers, and developers who paste their own code from elsewhere.

Flint uses a **multi-signal classifier** that combines 12 signals into a confidence vector across 4 session types. The classifier improves over time as it builds a personal baseline for each developer.

---

## The Four Session Types

### Human
The developer is writing the code themselves. Standard daemon behaviour. Stillness rule applies (90 seconds).

### Agent-assisted
The developer is writing with occasional AI help. Mix of human keystrokes and discrete agent-written chunks. Human is directing, agent is helping with specific pieces.

### Vibe coding
The agent is writing most of the code. Developer is directing, reviewing, and accepting. Large coherent chunks across multiple files. Human activity is primarily review and acceptance.

### Pair programming
Two humans writing together. Different keystroke rhythm, different style, possibly different timezone or hour-of-day pattern. Longer pauses between changes (discussion). Code quality is often higher or lower than this developer's solo baseline.

---

## The 12 Detection Signals

Each signal has a weight and a direction of evidence for each session type.

```
Signal                           Weight  Human  Agent  Vibe   Pair
────────────────────────────────────────────────────────────────────
1.  Keystroke cadence            0.20    ↑      ✗      ↓      ↑
2.  Typo + backspace rate        0.15    ↑      ✗      ✗      ↑
3.  Change size distribution     0.20    ↓      ↑      ↑      ↓
4.  Multi-file simultaneous      0.15    ↓      ↓      ↑      ↓
5.  Style consistency in chunk   0.10    ↓      ↑      ↑      ↓
6.  Comment density              0.08    ↓      ↑      ↑      ↓
7.  Variable name genericness    0.07    ↓      ↑      ↑      ↓
8.  Save pattern regularity      0.05    ↑      ↓      ↓      ↑
9.  Hour-of-day anomaly          0.05    ↓      ✗      ✗      ↑
10. Clipboard activity spike     0.05    ↓      ✗      ↑      ↓
11. Review pattern after write   0.05    ↓      ↓      ↑      ↓
12. Cursor movement pattern      0.05    ↑      ✗      ↓      ↑

↑ = strong evidence of this type
↓ = evidence against this type
✗ = not applicable / neutral
```

### Signal definitions

**1. Keystroke cadence (weight: 0.20)**
Humans type in irregular bursts with natural thinking pauses of 2–30 seconds. Agents produce no keystrokes before a large write. Pair programmers have a different rhythm from the primary developer — detectable after 3+ sessions of baseline.

Detection: measure inter-keystroke intervals over a 5-minute sliding window. High variance with 2–30 second gaps = human signal. Zero keystrokes before large file write = agent signal. Rhythm that doesn't match developer's baseline = pair signal.

**2. Typo + backspace rate (weight: 0.15)**
Humans mistype. Agents don't. Pair programmers also mistype but at a different rate and with different common errors than the primary developer.

Detection: count backspace keystrokes as a percentage of total keystrokes over last 500 keystrokes. > 3% = strong human signal. 0% over 50+ lines = strong agent signal. Different error pattern from baseline = pair signal.

**3. Change size distribution (weight: 0.20)**
Humans typically write in 1–20 line increments. Agents write in 20–200 line coherent chunks. Vibe coding shows multiple large chunks across files. Pair programmers write in human-sized increments but the distribution shifts when the other person takes over.

Detection: histogram of file change sizes (lines delta) over the session. Bimodal distribution (small + large) = agent-assisted. Consistently large = vibe coding. Human-sized but shifted from baseline = pair.

**4. Multi-file simultaneous change (weight: 0.15)**
Humans focus on one file at a time. Vibe coding produces changes across multiple files in rapid succession (< 5 seconds apart). Agents working on a single task produce changes in one file. Pair programmers change files sequentially.

Detection: count file changes within 5-second windows. More than 2 different files in 5 seconds = strong vibe coding signal.

**5. Style consistency within chunk (weight: 0.10)**
Agent-written code has unnaturally consistent style within a chunk — uniform indentation, consistent naming conventions, no variation. Human-written code has natural micro-variation. Vibe coding chunks are also agent-written so same signal.

Detection: measure naming convention consistency, indentation uniformity, and comment style within chunks > 20 lines. Unnaturally high consistency = agent signal.

**6. Comment density (weight: 0.08)**
Agents over-comment. They explain what every block does even when it's obvious. Humans comment selectively and often forget to comment at all. High comment density (> 1 comment per 5 lines) on new code = agent signal.

Detection: count comment lines as percentage of total new lines in a chunk. > 20% = agent signal. < 5% on complex code = human signal.

**7. Variable name genericness (weight: 0.07)**
Agents default to generic names: `data`, `result`, `item`, `temp`, `response`, `error`, `value`. Humans use domain-specific names. Count of generic variable names as percentage of all new variable names = agent signal.

Detection: compare new variable names against a list of 50 common agent-default names. > 30% generic = agent signal.

**8. Save pattern regularity (weight: 0.05)**
Humans save irregularly — when they remember, when something feels done, after fixing an error. Agents trigger saves programmatically — regular intervals or immediately after writing. Vibe coding saves are triggered by the agent's completion signal.

Detection: measure inter-save intervals. High regularity (low variance) = agent signal. Irregular with clustering around error fixes = human signal.

**9. Hour-of-day anomaly (weight: 0.05)**
After 2+ weeks of usage, Flint knows when this developer typically codes. Activity at significantly unusual hours is a pair signal — the other person is in a different timezone or working different hours.

Detection: compare current session hour against developer's historical activity distribution. More than 2 standard deviations from mean = pair signal.

**10. Clipboard activity spike (weight: 0.05)**
Vibe coding involves frequent clipboard activity — copying prompts to paste into agents, copying agent output to paste into the editor or terminal. Clipboard writes/reads spike during vibe coding sessions.

Detection: monitor clipboard change frequency (where OS permits). > 5 clipboard events per minute during a coding session = vibe coding signal.

**11. Review pattern after write (weight: 0.05)**
After an agent writes a chunk, the human typically reviews it — short cursor movements over the new code, occasional small edits, no new large writes for 10–30 seconds. This review pattern is distinct from normal coding.

Detection: after a large write (> 30 lines), look for short cursor activity and small edits (< 5 lines) in the 10–30 second window. Present = vibe coding signal.

**12. Cursor movement pattern (weight: 0.05)**
Humans re-read their code — cursor jumps up to check earlier lines, scrolls to related functions, returns. Agents don't move the cursor before writing. Pair programmers have a different cursor movement pattern from the primary developer.

Detection: measure cursor jump frequency and distance. Frequent short jumps = human signal. No cursor movement before large write = agent signal. Movement pattern that doesn't match developer baseline = pair signal.

---

## The Classifier

```go
// core/internal/daemon/session.go

type SessionType int

const (
    SessionHuman         SessionType = iota
    SessionAgentAssisted
    SessionVibeCoding
    SessionPairProgramming
    SessionMixed
)

type SignalVector struct {
    Human  float64
    Agent  float64
    Vibe   float64
    Pair   float64
}

type SessionClassifier struct {
    baseline    *DeveloperBaseline
    recentSignals []Signal
    window      time.Duration // 5-minute sliding window
}

func (c *SessionClassifier) Classify() (SessionType, float64) {
    vector := c.computeVector()
    max, secondMax := topTwo(vector)

    // If top two are within 0.10 — declare mixed
    if max.score - secondMax.score < 0.10 {
        return SessionMixed, max.score
    }

    return max.sessionType, max.score
}

func (c *SessionClassifier) computeVector() SignalVector {
    v := SignalVector{}

    signals := []struct {
        weight float64
        score  func() SignalVector
    }{
        {0.20, c.keystrokeCadenceSignal},
        {0.15, c.typoBackspaceSignal},
        {0.20, c.changeSizeSignal},
        {0.15, c.multiFileSignal},
        {0.10, c.styleConsistencySignal},
        {0.08, c.commentDensitySignal},
        {0.07, c.variableNameSignal},
        {0.05, c.savePatternSignal},
        {0.05, c.hourAnomalySignal},
        {0.05, c.clipboardSignal},
        {0.05, c.reviewPatternSignal},
        {0.05, c.cursorMovementSignal},
    }

    for _, s := range signals {
        sv := s.score()
        v.Human += s.weight * sv.Human
        v.Agent += s.weight * sv.Agent
        v.Vibe  += s.weight * sv.Vibe
        v.Pair  += s.weight * sv.Pair
    }

    return v
}
```

---

## Personal Baseline

The classifier improves over time by learning this developer's personal baseline.

```go
// core/internal/daemon/baseline.go

type DeveloperBaseline struct {
    // Built from first 10 confirmed-human sessions
    AvgKeystrokeCadence    time.Duration
    KeystrokeCadenceStdDev time.Duration
    AvgChangeSize          float64
    ChangeSizeStdDev       float64
    TypicalCodingHours     [24]float64 // probability distribution
    AvgSaveInterval        time.Duration
    TypicalCommentDensity  float64
    CommonVariableNames    map[string]int // this developer's actual naming patterns
    BaselineConfidence     float64 // 0-1, how confident baseline is (needs 10+ sessions)
}

// After 10+ confirmed-human sessions, signals use deviation from baseline
// instead of absolute values — dramatically improves accuracy
func (c *SessionClassifier) keystrokeCadenceSignal() SignalVector {
    cadence := c.measureCadence()

    if c.baseline.BaselineConfidence < 0.7 {
        // Not enough data — use absolute thresholds
        return absoluteKeystrokeSignal(cadence)
    }

    // Use deviation from personal baseline
    deviations := (cadence - c.baseline.AvgKeystrokeCadence) / c.baseline.KeystrokeCadenceStdDev
    return deviationKeystrokeSignal(deviations)
}
```

---

## Agent Fingerprinting

After enough sessions, Flint can distinguish between different AI agents by their output patterns.

```go
// core/internal/daemon/agent_fingerprint.go

type AgentFingerprint struct {
    Name              string
    CommentStyle      string  // "// descriptive" vs "// inline" vs JSDoc
    VariableNaming    string  // camelCase preference, length distribution
    ErrorHandlingStyle string // try/catch vs if-err vs .catch()
    TestWritingStyle  string  // describe/it vs test() vs none
    BoilerplatePatterns []string // common opening patterns
}

// Known agent fingerprints (updated via shared/agent_fingerprints.json)
var KnownAgents = []AgentFingerprint{
    {
        Name: "Cursor",
        CommentStyle: "inline_sparse",
        VariableNaming: "descriptive_camel",
        ErrorHandlingStyle: "explicit_try_catch",
    },
    {
        Name: "Antigravity",
        CommentStyle: "block_heavy",
        VariableNaming: "verbose_descriptive",
        ErrorHandlingStyle: "defensive_null_check_first",
    },
    // Community-contributed fingerprints loaded from shared/agent_fingerprints.json
}

func (f *AgentFingerprintDetector) Detect(chunk *CodeChunk) (string, float64) {
    scores := map[string]float64{}
    for _, agent := range KnownAgents {
        scores[agent.Name] = agent.ScoreChunk(chunk)
    }
    return maxScore(scores)
}
```

---

## Pair Programming Detection — Special Handling

When pair programming is detected, Flint adjusts its behaviour:

- Observation threshold raised slightly — two people are already reviewing each other's work
- Human intelligence observations suppressed — behavioural signals are unreliable with two people
- Win acknowledgments adjusted — "you two cleaned that up nicely" vs solo
- In v4 team mode: session attributed to pair, not individual
- Knowledge transfer flag set — this area of the codebase now has two people who understand it

---

## Detection Accuracy Targets

| Session type | Target accuracy (after baseline) | Target accuracy (cold start) |
|---|---|---|
| Human | 95% | 85% |
| Agent-assisted | 90% | 75% |
| Vibe coding | 92% | 82% |
| Pair programming | 85% | 65% |
| Mixed | 80% | 70% |

Cold start (first 10 sessions) accuracy is lower — this is acceptable because early observations are held to a higher confidence threshold (0.85 instead of 0.75) until baseline is established.

---

## Privacy Considerations

**Keystroke monitoring:** Flint counts keystrokes and measures cadence. It never records which keys were pressed — only the timing and count. No content is captured.

**Cursor movement:** Flint tracks cursor position changes in the editor (via VS Code API). It never records which line or column — only the frequency and distance of jumps.

**Clipboard:** Clipboard monitoring is opt-in, off by default, and only measures change frequency — never content.

**Pair detection:** When pair programming is detected, the observation is stored as "session type: pair" with no attempt to identify the second person. Their identity is never inferred or stored.

All of this is covered by the standard `flint memory` transparency and `flint memory --clear` deletion commands.

---
