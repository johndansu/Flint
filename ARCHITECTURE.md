# Architecture Document
## Flint — Complete System Architecture, Workflows & Skeletons

**Version:** 1.0
**Status:** Draft
**Last updated:** May 2026
**Scope:** All versions (v1 → v4)

---

## 1. Philosophy

Flint is built on three architectural principles:

**Local first.** Every byte of data stays on the developer's machine. The cloud exists only to run AI inference — never to store, log, or process developer data.

**Go for runtime, TypeScript for interface.** The daemon and CLI are Go — compiled, fast, low memory, built for long-running processes. The editor extension is TypeScript — the only real option for VS Code API. They communicate through a shared IPC schema. One repo, one pipeline, two runtimes that never need to know about each other's internals.

**One schema, one contract.** Every message between Go and TypeScript is defined in a single JSON schema. Both sides are generated from it. If the schema changes, both sides break at compile time, not at runtime.

---

## 2. Monorepo Structure

```
flint/
├── Makefile                    # single entry point for all build commands
├── turbo.json                  # turborepo pipeline config
├── .github/
│   └── workflows/
│       ├── ci.yml              # lint, test, build on every PR
│       └── release.yml         # cross-compile + publish on tag
│
├── core/                       # Go — daemon + CLI
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/
│   │   ├── flint/              # CLI entry point
│   │   │   └── main.go
│   │   └── daemon/             # daemon entry point
│   │       └── main.go
│   ├── internal/
│   │   ├── awareness/          # Spark / Flame / Forge logic
│   │   │   └── awareness.go
│   │   ├── daemon/
│   │   │   ├── daemon.go       # daemon lifecycle
│   │   │   ├── watcher.go      # file system watcher (fsnotify)
│   │   │   ├── git.go          # git activity monitor
│   │   │   ├── deps.go         # dependency CVE + health checker
│   │   │   ├── session.go      # session type + state detection
│   │   │   ├── classifier.go   # 12-signal multi-factor classifier
│   │   │   ├── baseline.go     # personal developer baseline builder
│   │   │   ├── fingerprint.go  # agent fingerprint detector
│   │   │   ├── technical.go    # technical observation engine
│   │   │   ├── human.go        # human intelligence engine
│   │   │   ├── footguns.go     # footgun library loader + pattern matcher
│   │   │   ├── watch.go        # tripwire engine
│   │   │   ├── scanner.go      # onboarding scan (flint scan)
│   │   │   ├── calibration.go  # per-category threshold management
│   │   │   └── timegate.go     # time-gating rules
│   │   ├── nudge/
│   │   │   └── nudge.go        # CLI cross-promotion engine
│   │   ├── ipc/
│   │   │   ├── server.go       # Unix socket server
│   │   │   ├── messages.go     # message types (generated from schema)
│   │   │   └── broadcast.go    # broadcast to all connected clients
│   │   ├── store/
│   │   │   ├── store.go        # SQLite wrapper
│   │   │   ├── observations.go # observations store
│   │   │   ├── errors.go       # error dictionary store
│   │   │   ├── human.go        # human intelligence store
│   │   │   └── calibration.go  # calibration store
│   │   ├── ai/
│   │   │   ├── client.go       # Anthropic API client
│   │   │   ├── prompts.go      # all system prompts
│   │   │   └── stream.go       # streaming response handler
│   │   ├── config/
│   │   │   └── config.go       # ~/.flint/config.json reader/writer
│   │   └── tools/              # 24 manual tools + new v1 tools
│   │       ├── tools.go        # tool registry (loaded from shared/tools.json)
│   │       ├── review.go
│   │       ├── debug.go
│   │       ├── doc.go
│   │       ├── audit.go
│   │       ├── explain.go      # flint explain
│   │       └── diff.go         # flint diff
│   └── pkg/
│       └── schema/             # shared types (mirrored in TS)
│           └── schema.go
│
├── extension/                  # TypeScript — VS Code / Cursor / Windsurf
│   ├── package.json
│   ├── tsconfig.json
│   ├── src/
│   │   ├── extension.ts        # extension entry point
│   │   ├── ipc/
│   │   │   ├── client.ts       # Unix socket client
│   │   │   └── messages.ts     # message types (generated from schema)
│   │   ├── sidebar/
│   │   │   ├── provider.ts     # sidebar webview provider
│   │   │   ├── card.ts         # observation card renderer
│   │   │   └── quiet.ts        # quiet state (4px strip)
│   │   ├── agent/
│   │   │   ├── cursor.ts       # Cursor agent integration
│   │   │   ├── antigravity.ts  # Antigravity integration
│   │   │   ├── windsurf.ts     # Windsurf integration
│   │   │   └── copilot.ts      # GitHub Copilot integration
│   │   ├── commands/           # VS Code command palette entries
│   │   │   └── commands.ts
│   │   └── status/
│   │       └── bar.ts          # status bar item (awareness level)
│   └── media/
│       ├── sidebar.css
│       └── icon.svg
│
├── wrapper/                    # TypeScript — npm install wrapper
│   ├── package.json            # this is what `npm install -g flint` installs
│   ├── install.js              # post-install: detect OS/arch, download binary
│   ├── bin/
│   │   └── flint.js            # thin shim that exec's the Go binary
│   └── releases/               # downloaded binaries land here (gitignored)
│
├── shared/
│   ├── schema.json             # THE contract — all IPC message types
│   ├── tools.json              # tool definitions (role + category + prompts)
│   ├── footguns.json           # stack-aware footgun library (community-contributed)
│   └── agent_fingerprints.json # known agent output fingerprints
│
├── scripts/
│   ├── generate-schema.js      # generates Go + TS types from schema.json
│   ├── cross-compile.sh        # builds all platform binaries
│   └── publish.sh              # publishes npm wrapper + VS Code extension
│
└── docs/
    ├── PRD.md
    ├── MVP.md
    ├── ARCHITECTURE.md         # this file
    ├── PITCH.md
    ├── GO_TO_MARKET.md
    ├── CONTRIBUTING.md
    └── CHANGELOG.md
```

---

## 3. The Shared Schema — The Contract

`shared/schema.json` is the single source of truth for all communication between Go and TypeScript. Generated types are never edited manually.

```json
{
  "$schema": "http://json-schema.org/draft-07/schema",
  "messages": {
    "observation": {
      "type": "object",
      "properties": {
        "id":         { "type": "string" },
        "kind":       { "enum": ["technical", "human", "win", "question", "taste"] },
        "category":   { "type": "string" },
        "confidence": { "type": "number", "minimum": 0, "maximum": 1 },
        "text":       { "type": "string" },
        "agentPrompt":{ "type": "string" },
        "timestamp":  { "type": "string", "format": "date-time" },
        "filePath":   { "type": "string" },
        "lineNumber":  { "type": "integer" }
      },
      "required": ["id", "kind", "category", "confidence", "text", "timestamp"]
    },
    "dismissal": {
      "type": "object",
      "properties": {
        "observationId": { "type": "string" },
        "timestamp":     { "type": "string", "format": "date-time" }
      }
    },
    "sessionState": {
      "type": "object",
      "properties": {
        "state": { "enum": ["exploration", "deepWork", "debuggingSpiral", "finishing", "early"] },
        "since": { "type": "string", "format": "date-time" }
      }
    },
    "daemonStatus": {
      "type": "object",
      "properties": {
        "awareness":       { "enum": ["spark", "flame", "forge"] },
        "humanLayer":      { "type": "boolean" },
        "project":         { "type": "string" },
        "observationCount":{ "type": "integer" },
        "sessionState":    { "type": "string" },
        "lastObservation": { "type": "string", "format": "date-time" }
      }
    }
  }
}
```

Generate types:
```bash
node scripts/generate-schema.js
# outputs:
#   core/pkg/schema/schema.go
#   extension/src/ipc/messages.ts
```

---

## 4. IPC Architecture — Go ↔ TypeScript

Communication is over a Unix domain socket (named pipe on Windows).

```
Socket path:  ~/.flint/daemon.sock
Protocol:     newline-delimited JSON
Direction:    Go daemon → TypeScript extension (one-way broadcast)
              TypeScript extension → Go daemon (dismissals, config changes)
```

### Message flow

```
Go Daemon                           TypeScript Extension
─────────                           ────────────────────
daemon.sock (server)    ←connect→   ipc/client.ts
                        ←────────   { type: "subscribe" }
observation generated
                        ────────→   { type: "observation", ...data }
                                    card renders in sidebar
                                    developer taps ✕
                        ←────────   { type: "dismissal", observationId }
calibration updated
                        ────────→   { type: "ack" }
```

### Go IPC server skeleton

```go
// core/internal/ipc/server.go
package ipc

import (
    "bufio"
    "encoding/json"
    "net"
    "os"
    "sync"
)

type Server struct {
    socketPath string
    clients    map[net.Conn]bool
    mu         sync.RWMutex
    onMessage  func(msg Message)
}

func NewServer(socketPath string, onMessage func(msg Message)) *Server {
    return &Server{
        socketPath: socketPath,
        clients:    make(map[net.Conn]bool),
        onMessage:  onMessage,
    }
}

func (s *Server) Start() error {
    os.Remove(s.socketPath)
    ln, err := net.Listen("unix", s.socketPath)
    if err != nil {
        return err
    }
    go s.acceptLoop(ln)
    return nil
}

func (s *Server) Broadcast(msg Message) {
    data, _ := json.Marshal(msg)
    data = append(data, '\n')
    s.mu.RLock()
    defer s.mu.RUnlock()
    for conn := range s.clients {
        conn.Write(data)
    }
}

func (s *Server) acceptLoop(ln net.Listener) {
    for {
        conn, err := ln.Accept()
        if err != nil { continue }
        s.mu.Lock()
        s.clients[conn] = true
        s.mu.Unlock()
        go s.readLoop(conn)
    }
}

func (s *Server) readLoop(conn net.Conn) {
    defer func() {
        s.mu.Lock()
        delete(s.clients, conn)
        s.mu.Unlock()
        conn.Close()
    }()
    scanner := bufio.NewScanner(conn)
    for scanner.Scan() {
        var msg Message
        if err := json.Unmarshal(scanner.Bytes(), &msg); err == nil {
            s.onMessage(msg)
        }
    }
}
```

### TypeScript IPC client skeleton

```typescript
// extension/src/ipc/client.ts
import * as net from 'net';
import * as os from 'os';
import * as path from 'path';
import { Message, Observation } from './messages';

export class IPCClient {
  private socket: net.Socket | null = null;
  private socketPath = path.join(os.homedir(), '.flint', 'daemon.sock');
  private onObservation: (obs: Observation) => void;
  private buffer = '';

  constructor(onObservation: (obs: Observation) => void) {
    this.onObservation = onObservation;
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.socket = net.createConnection(this.socketPath, () => {
        this.send({ type: 'subscribe' });
        resolve();
      });
      this.socket.on('data', (data) => this.handleData(data.toString()));
      this.socket.on('error', reject);
      this.socket.on('close', () => setTimeout(() => this.connect(), 3000));
    });
  }

  private handleData(data: string) {
    this.buffer += data;
    const lines = this.buffer.split('\n');
    this.buffer = lines.pop() || '';
    for (const line of lines) {
      if (!line.trim()) continue;
      try {
        const msg: Message = JSON.parse(line);
        if (msg.type === 'observation') {
          this.onObservation(msg.data as Observation);
        }
      } catch {}
    }
  }

  send(msg: object) {
    this.socket?.write(JSON.stringify(msg) + '\n');
  }

  dismiss(observationId: string) {
    this.send({ type: 'dismissal', observationId });
  }
}
```

---

## 5. Daemon Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      Flint Daemon                        │
│                                                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│  │  Watcher │  │   Git    │  │   Deps   │              │
│  │ (fsnotify)│  │ Monitor  │  │  Checker │              │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘              │
│       │              │              │                    │
│       └──────────────┴──────────────┘                   │
│                       │                                  │
│              ┌────────▼────────┐                        │
│              │  Session State  │                        │
│              │    Detector     │                        │
│              └────────┬────────┘                        │
│                       │                                  │
│         ┌─────────────┼─────────────┐                  │
│         │             │             │                    │
│  ┌──────▼──────┐ ┌───▼────┐ ┌─────▼──────┐            │
│  │  Technical  │ │  Human │ │  Time Gate │            │
│  │   Engine    │ │ Engine │ │   Guard    │            │
│  └──────┬──────┘ └───┬────┘ └─────┬──────┘            │
│         │             │            │                    │
│         └─────────────┴────────────┘                   │
│                       │                                  │
│              ┌────────▼────────┐                        │
│              │  Observation    │                        │
│              │    Scorer       │  (confidence 0-1)      │
│              └────────┬────────┘                        │
│                       │                                  │
│              ┌────────▼────────┐                        │
│              │  IPC Broadcast  │──────→ extension       │
│              └────────┬────────┘        terminal        │
│                       │                                  │
│              ┌────────▼────────┐                        │
│              │   SQLite Store  │                        │
│              └─────────────────┘                        │
└─────────────────────────────────────────────────────────┘
```

### Daemon skeleton

```go
// core/internal/daemon/daemon.go
package daemon

import (
    "context"
    "path/filepath"
    "time"
)

type Daemon struct {
    config     *Config
    store      *Store
    ipc        *IPCServer
    watcher    *Watcher
    git        *GitMonitor
    deps       *DepsChecker
    session    *SessionDetector
    technical  *TechnicalEngine
    human      *HumanEngine
    timegate   *TimeGate
    calibration *CalibrationStore
}

func New(config *Config) *Daemon {
    store := NewStore(filepath.Join(config.FlintDir, "flint.db"))
    return &Daemon{
        config:      config,
        store:       store,
        ipc:         NewIPCServer(config),
        watcher:     NewWatcher(),
        git:         NewGitMonitor(),
        deps:        NewDepsChecker(),
        session:     NewSessionDetector(),
        technical:   NewTechnicalEngine(store),
        human:       NewHumanEngine(store),
        timegate:    NewTimeGate(store),
        calibration: NewCalibrationStore(store),
    }
}

func (d *Daemon) Start(ctx context.Context) error {
    if err := d.ipc.Start(); err != nil {
        return err
    }
    if err := d.watcher.Watch(d.config.ProjectPath); err != nil {
        return err
    }
    go d.loop(ctx)
    return nil
}

func (d *Daemon) loop(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case event := <-d.watcher.Events():
            d.session.RecordActivity(event)
        case <-ticker.C:
            d.evaluate()
        }
    }
}

func (d *Daemon) evaluate() {
    state := d.session.CurrentState()
    sessionType := d.session.CurrentType() // human | agentAssisted | vibeCoding | mixed

    // Never evaluate during early session
    if state == SessionEarly { return }

    // Check time gates — different limits per session type
    if !d.timegate.CanSpeak(sessionType) { return }

    // Vibe coding mode — completely different evaluation path
    if sessionType == SessionVibeCoding {
        if obs := d.technical.EvaluateAgentOutput(state); obs != nil {
            d.maybeEmit(obs)
        }
        return
    }

    // Human / agent-assisted — wait for stillness (90s)
    if !d.session.IsStill() { return }

    // Run technical engine
    if d.config.Awareness >= AwareFlame {
        if obs := d.technical.Evaluate(state); obs != nil {
            d.maybeEmit(obs)
        }
    }

    // Run human engine (opt-in)
    if d.config.HumanLayer && d.config.Awareness >= AwareForge {
        if obs := d.human.Evaluate(state); obs != nil {
            d.maybeEmit(obs)
        }
    }
}

func (d *Daemon) maybeEmit(obs *Observation) {
    threshold := d.calibration.ThresholdFor(obs.Category)
    if obs.Confidence < threshold { return }

    d.store.SaveObservation(obs)
    d.timegate.RecordEmission(obs.Category)
    d.ipc.Broadcast(obs)
}
```

---

## 5b. Vibe Coding Detection — Session Type Architecture

The session detector is the most important component in the daemon. It determines which evaluation path fires — everything else depends on it getting this right.

### Session types

```go
// core/internal/daemon/session.go

type SessionType int

const (
    SessionHuman         SessionType = iota // human typing, gradual changes
    SessionAgentAssisted                    // human + occasional agent help
    SessionVibeCoding                       // agent writing, human directing
    SessionMixed                            // transitioned mid-session
)

type SessionState int

const (
    SessionEarly         SessionState = iota // first 15 minutes
    SessionExploration                       // rapid file switching
    SessionDeepWork                          // focused, steady progress
    SessionDebuggingSpiral                   // repeated reverts, same lines
    SessionFinishing                         // small cleanup changes
)

type SessionDetector struct {
    recentChanges    []FileChange
    keystrokeTracker *KeystrokeTracker
    currentType      SessionType
    currentState     SessionState
    sessionStart     time.Time
    lastLargeWrite   time.Time
    mu               sync.RWMutex
}

type FileChange struct {
    Path      string
    LinesAdded   int
    LinesRemoved int
    Timestamp time.Time
    HasKeystrokes bool // were keystrokes detected before this change?
}

func (s *SessionDetector) RecordChange(change FileChange) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.recentChanges = append(s.recentChanges, change)
    s.pruneOldChanges()
    s.updateType()
    s.updateState()
}

func (s *SessionDetector) updateType() {
    recent := s.changesInLast(5 * time.Minute)

    // Detect vibe coding: large changes, no preceding keystrokes
    largeAgentWrites := 0
    humanChanges := 0
    for _, c := range recent {
        if c.LinesAdded > 50 && !c.HasKeystrokes {
            largeAgentWrites++
        } else if c.HasKeystrokes {
            humanChanges++
        }
    }

    if largeAgentWrites > 2 && humanChanges < 2 {
        if s.currentType == SessionHuman || s.currentType == SessionAgentAssisted {
            s.currentType = SessionMixed // transitioned
        } else {
            s.currentType = SessionVibeCoding
        }
        return
    }

    if largeAgentWrites > 0 && humanChanges > 0 {
        s.currentType = SessionAgentAssisted
        return
    }

    if s.currentType == SessionVibeCoding && humanChanges > largeAgentWrites {
        s.currentType = SessionMixed
        return
    }

    s.currentType = SessionHuman
}

// IsAgentChunkComplete returns true when a vibe coding chunk is done
// "done" = large write followed by 10 seconds of no file activity
func (s *SessionDetector) IsAgentChunkComplete() bool {
    if s.currentType != SessionVibeCoding && s.currentType != SessionMixed {
        return false
    }
    lastChange := s.lastChangeTime()
    return time.Since(lastChange) > 10*time.Second &&
           time.Since(s.lastLargeWrite) < 2*time.Minute
}

// WasAcceptedQuickly returns true when a large change was accepted
// with less than 30 seconds of review activity
func (s *SessionDetector) WasAcceptedQuickly() bool {
    if s.lastLargeWrite.IsZero() { return false }
    timeSinceWrite := time.Since(s.lastLargeWrite)
    return timeSinceWrite < 30*time.Second
}
```

### Vibe coding evaluator

```go
// core/internal/daemon/technical.go (vibe coding path)

func (e *TechnicalEngine) EvaluateAgentOutput(state SessionState) *Observation {
    // Get the most recent large agent-written chunk
    chunk := e.store.GetLastAgentChunk()
    if chunk == nil { return nil }

    // Check each vibe coding observation category
    checks := []func(*CodeChunk) *Observation{
        e.checkErrorHandling,
        e.checkEdgeCases,
        e.checkSecurityPatterns,
        e.checkCodebaseConsistency,
        e.checkOverEngineering,
        e.checkAcceptanceRisk,
    }

    for _, check := range checks {
        if obs := check(chunk); obs != nil {
            return obs
        }
    }
    return nil
}

func (e *TechnicalEngine) checkAcceptanceRisk(chunk *CodeChunk) *Observation {
    if !e.session.WasAcceptedQuickly() { return nil }
    if chunk.LinesAdded < 100 { return nil }

    // Find top 3 things to check using AI
    analysis := e.ai.AnalyzeQuickAcceptance(chunk)
    return &Observation{
        Kind:       "technical",
        Category:   "acceptance_risk",
        Confidence: 0.90,
        Text:       analysis.Summary,
        AgentPrompt: analysis.ReviewPrompt,
    }
}
```

### Time gate adjustments per session type

```go
// core/internal/daemon/timegate.go

func (t *TimeGate) CanSpeak(sessionType SessionType) bool {
    switch sessionType {
    case SessionVibeCoding, SessionMixed:
        // Shorter minimum gap — agent output arrives faster
        return t.timeSinceLastEmission() > 10*time.Minute &&
               t.emissionsThisSession() < 8
    default:
        // Standard human coding limits
        return t.timeSinceLastEmission() > 20*time.Minute &&
               t.emissionsThisSession() < 5
    }
}
```

---

## 6. CLI Architecture

### Command structure

```
flint
├── init                 # first run wizard
├── start                # start daemon
├── stop                 # stop daemon
├── status               # daemon status + awareness
├── config               # update role, awareness, API key, agent
├── awareness            # change awareness level
├── human
│   ├── on               # enable human layer
│   ├── off              # disable human layer
│   ├── status           # human layer status
│   └── --clear          # delete human data
├── review               # code review tool
├── debug                # error explainer
├── doc                  # documentation generator
├── audit                # security / performance audit
├── check                # manual pre-commit check
├── update               # stakeholder translation
│   ├── --commits N
│   ├── --format [slack|email|plain]
│   └── --audience [dev|pm|cto]
└── memory
    ├── (default)        # show all stored data
    ├── --clear          # delete everything
    └── --export         # export as JSON
```

### CLI skeleton

```go
// core/cmd/flint/main.go
package main

import (
    "os"
    "github.com/spf13/cobra"
)

func main() {
    root := &cobra.Command{
        Use:   "flint",
        Short: "Passive technical and human intelligence for developers",
        Run:   runInteractive, // no args = interactive picker
    }

    root.AddCommand(
        newInitCmd(),
        newStartCmd(),
        newStopCmd(),
        newStatusCmd(),
        newConfigCmd(),
        newAwarenessCmd(),
        newHumanCmd(),
        newReviewCmd(),
        newDebugCmd(),
        newDocCmd(),
        newAuditCmd(),
        newCheckCmd(),
        newUpdateCmd(),
        newMemoryCmd(),
    )

    if err := root.Execute(); err != nil {
        os.Exit(1)
    }
}
```

### Tool command skeleton

```go
// core/cmd/flint/review.go
package main

import (
    "fmt"
    "io"
    "os"
    "github.com/spf13/cobra"
    "github.com/flint/internal/ai"
    "github.com/flint/internal/config"
    "github.com/flint/internal/tools"
)

func newReviewCmd() *cobra.Command {
    var role string

    cmd := &cobra.Command{
        Use:   "review",
        Short: "Code review",
        RunE: func(cmd *cobra.Command, args []string) error {
            cfg := config.Load()
            if role == "" { role = cfg.Role }

            input, err := io.ReadAll(os.Stdin)
            if err != nil { return err }

            tool := tools.Get("review", role)
            if err := ai.Stream(cfg.APIKey, tool.Prompt, string(input), os.Stdout); err != nil {
                return err
            }

            // Cross-promotion: nudge toward extension if running manually often
            nudge.MaybeNudgeExtension(cfg, "review")
            return nil
        },
    }

    cmd.Flags().StringVarP(&role, "role", "r", "", "override configured role")
    return cmd
}
```

### CLI cross-promotion engine

```go
// core/internal/nudge/nudge.go
package nudge

import (
    "fmt"
    "time"
    "github.com/flint/internal/config"
    "github.com/flint/internal/store"
)

type NudgeType string

const (
    NudgeExtensionRepeatedTool   NudgeType = "extension_repeated_tool"
    NudgeExtensionManualCheck    NudgeType = "extension_manual_check"
    NudgeExtensionRepeatedDebug  NudgeType = "extension_repeated_debug"
)

// MaybeNudgeExtension checks if a contextual nudge should fire
// and prints it to stdout if so. Never interrupts — always appears
// after the tool output, separated by a blank line.
func MaybeNudgeExtension(cfg *config.Config, toolName string) {
    if cfg.NudgeSuppressed { return }

    s := store.Open()
    count := s.ToolUsageToday(toolName)
    lastNudge := s.LastNudge(string(NudgeExtensionRepeatedTool))

    // Only nudge if used 3+ times today and not nudged this type today
    if count < 3 { return }
    if time.Since(lastNudge) < 24*time.Hour { return }

    s.RecordNudge(string(NudgeExtensionRepeatedTool))

    fmt.Println()
    fmt.Println("  💡 You'd get this automatically in your editor.")
    fmt.Println("     Install the Flint extension:")
    fmt.Println("     https://marketplace.visualstudio.com/items?itemName=flint-dev.flint")
    fmt.Println("     (flint config --no-nudge to suppress)")
}

func MaybeNudgeExtensionCheck(cfg *config.Config) {
    if cfg.NudgeSuppressed { return }

    s := store.Open()
    lastNudge := s.LastNudge(string(NudgeExtensionManualCheck))
    if time.Since(lastNudge) < 24*time.Hour { return }

    s.RecordNudge(string(NudgeExtensionManualCheck))

    fmt.Println()
    fmt.Println("  💡 The extension runs this automatically on every commit.")
    fmt.Println("     https://marketplace.visualstudio.com/items?itemName=flint-dev.flint")
}
```

---

## 7. Extension Architecture

### CLI resolution — which Flint does the extension use?

```typescript
// extension/src/cli/resolver.ts
import * as vscode from 'vscode';
import * as path from 'path';
import * as fs from 'fs';
import { execSync } from 'child_process';
import * as os from 'os';

export type CLIResolution =
  | { found: true;  path: string; version: string; source: 'local' | 'config' | 'global' | 'path' }
  | { found: false; reason: string };

export async function resolveCLI(): Promise<CLIResolution> {
    const workspaceRoot = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;

    // 1. Workspace-local install
    if (workspaceRoot) {
        const local = path.join(workspaceRoot, 'node_modules', '.bin', 'flint');
        if (fs.existsSync(local)) return probe(local, 'local');
    }

    // 2. .flintrc cliPath override
    if (workspaceRoot) {
        const rc = path.join(workspaceRoot, '.flintrc');
        if (fs.existsSync(rc)) {
            try {
                const config = JSON.parse(fs.readFileSync(rc, 'utf8'));
                if (config.cliPath && fs.existsSync(config.cliPath)) {
                    return probe(config.cliPath, 'config');
                }
            } catch {}
        }
    }

    // 3. Global npm install
    try {
        const npmRoot = execSync('npm root -g', { encoding: 'utf8' }).trim();
        const global = path.join(npmRoot, '..', 'bin', 'flint');
        if (fs.existsSync(global)) return probe(global, 'global');
    } catch {}

    // 4. PATH resolution
    try {
        const which = execSync('which flint', { encoding: 'utf8' }).trim();
        if (which) return probe(which, 'path');
    } catch {}

    // 5. Not found — standalone mode
    return { found: false, reason: 'Flint CLI not found. Install with: npm install -g flint' };
}

function probe(cliPath: string, source: CLIResolution['source'] & string): CLIResolution {
    try {
        const version = execSync(`${cliPath} --version`, { encoding: 'utf8' }).trim();
        return { found: true, path: cliPath, version, source: source as any };
    } catch {
        return { found: false, reason: `Found at ${cliPath} but failed to run` };
    }
}
```

### Per-workspace socket path

```typescript
// extension/src/ipc/socket.ts
import * as crypto from 'crypto';
import * as path from 'path';
import * as os from 'os';
import * as vscode from 'vscode';

export function getSocketPath(): string {
    const workspaceRoot = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath ?? 'default';
    const hash = crypto.createHash('md5').update(workspaceRoot).digest('hex').slice(0, 8);
    return path.join(os.homedir(), '.flint', `daemon-${hash}.sock`);
}
```

### Extension operating modes

```typescript
// extension/src/mode.ts
export type ExtensionMode =
    | 'connected'     // daemon alive, full experience
    | 'reconnecting'  // daemon was connected, socket dropped, retrying
    | 'standalone'    // no daemon found, Spark-level manual tools only
    | 'degraded';     // daemon found but version mismatch

export class ModeManager {
    private mode: ExtensionMode = 'standalone';
    private onModeChange: (mode: ExtensionMode) => void;

    constructor(onModeChange: (mode: ExtensionMode) => void) {
        this.onModeChange = onModeChange;
    }

    setMode(mode: ExtensionMode) {
        if (this.mode === mode) return;
        this.mode = mode;
        this.onModeChange(mode);
    }

    get current() { return this.mode; }

    isFullyFunctional() {
        return this.mode === 'connected';
    }

    isManualToolsAvailable() {
        // Manual tools work in all modes — direct API calls, no daemon needed
        return true;
    }

    isPassiveIntelligenceAvailable() {
        return this.mode === 'connected';
    }
}
```

### Extension entry point — full mode lifecycle

```typescript
// extension/src/extension.ts
import * as vscode from 'vscode';
import { resolveCLI } from './cli/resolver';
import { getSocketPath } from './ipc/socket';
import { IPCClient } from './ipc/client';
import { ModeManager } from './mode';
import { SidebarProvider } from './sidebar/provider';
import { StatusBar } from './status/bar';
import { CrossPromotion } from './nudge/cross-promotion';
import { registerCommands } from './commands/commands';

export async function activate(context: vscode.ExtensionContext) {
    const statusBar = new StatusBar();
    const modeManager = new ModeManager((mode) => statusBar.setMode(mode));
    const sidebar = new SidebarProvider(context, modeManager);
    const nudge = new CrossPromotion(context);

    // Resolve CLI
    const cli = await resolveCLI();
    if (!cli.found) {
        modeManager.setMode('standalone');
        statusBar.setCLINotFound();
        // Don't nudge immediately — wait for contextual trigger
    } else {
        statusBar.setCLI(cli.version, cli.source);

        // Connect to daemon
        const socketPath = getSocketPath();
        const ipc = new IPCClient(socketPath, {
            onObservation: (obs) => sidebar.showObservation(obs),
            onConnect: () => modeManager.setMode('connected'),
            onDisconnect: () => modeManager.setMode('reconnecting'),
            onVersionMismatch: () => modeManager.setMode('degraded'),
        });

        ipc.connect();
        context.subscriptions.push({ dispose: () => ipc.disconnect() });
    }

    context.subscriptions.push(
        vscode.window.registerWebviewViewProvider('flint.sidebar', sidebar),
        statusBar,
        nudge,
        ...registerCommands(context, { modeManager, sidebar }),
    );

    // Re-resolve CLI when workspace changes
    context.subscriptions.push(
        vscode.workspace.onDidChangeWorkspaceFolders(async () => {
            await activate(context); // re-run resolution
        })
    );
}

export function deactivate() {}
```

### Cross-promotion engine

```typescript
// extension/src/nudge/cross-promotion.ts
import * as vscode from 'vscode';

type NudgeType =
    | 'standalone_active_session'
    | 'flame_feature_needs_daemon'
    | 'new_project_no_flintrc';

export class CrossPromotion implements vscode.Disposable {
    private shownToday = new Set<NudgeType>();
    private sessionActiveMinutes = 0;
    private timer: NodeJS.Timeout;
    private suppressed: boolean;

    constructor(private context: vscode.ExtensionContext) {
        this.suppressed = context.globalState.get('flint.nudgeSuppressed', false);
        // Track active session time
        this.timer = setInterval(() => this.sessionActiveMinutes++, 60_000);
    }

    // Called by mode manager when standalone for 10+ minutes
    maybeNudgeStandaloneActive() {
        if (this.suppressed) return;
        if (this.shownToday.has('standalone_active_session')) return;
        if (this.sessionActiveMinutes < 10) return;

        this.show(
            'standalone_active_session',
            'Install the Flint CLI for passive intelligence',
            'npm install -g flint && flint init',
            'https://github.com/flint-dev/flint'
        );
    }

    // Called when developer selects a Flame/Forge feature without daemon
    nudgeFlameFeature(featureName: string) {
        if (this.suppressed) return;
        if (this.shownToday.has('flame_feature_needs_daemon')) return;

        this.show(
            'flame_feature_needs_daemon',
            `${featureName} requires the Flint CLI daemon`,
            'npm install -g flint && flint init',
            null
        );
    }

    // Called when new project opened with no .flintrc
    nudgeNewProject() {
        if (this.suppressed) return;
        if (this.shownToday.has('new_project_no_flintrc')) return;

        this.show(
            'new_project_no_flintrc',
            'Run `flint init` to enable full intelligence for this project',
            'flint init',
            null
        );
    }

    private show(type: NudgeType, message: string, command: string, link: string | null) {
        this.shownToday.add(type);
        const items: string[] = ['Copy command', "Don't show again"];
        if (link) items.unshift('Learn more');

        vscode.window.showInformationMessage(`Flint: ${message}`, ...items)
            .then(action => {
                if (action === 'Copy command') {
                    vscode.env.clipboard.writeText(command);
                }
                if (action === "Don't show again") {
                    this.suppressed = true;
                    this.context.globalState.update('flint.nudgeSuppressed', true);
                }
                if (action === 'Learn more' && link) {
                    vscode.env.openExternal(vscode.Uri.parse(link));
                }
            });
    }

    // Reset daily shown set at midnight
    private resetDaily() {
        const now = new Date();
        const midnight = new Date(now);
        midnight.setHours(24, 0, 0, 0);
        setTimeout(() => {
            this.shownToday.clear();
            this.resetDaily();
        }, midnight.getTime() - now.getTime());
    }

    dispose() { clearInterval(this.timer); }
}
```

### Status bar — all modes

```typescript
// extension/src/status/bar.ts
import * as vscode from 'vscode';
import { ExtensionMode } from '../mode';

export class StatusBar implements vscode.Disposable {
    private item: vscode.StatusBarItem;

    constructor() {
        this.item = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
        this.item.command = 'flint.openPanel';
        this.item.show();
    }

    setMode(mode: ExtensionMode) {
        const labels: Record<ExtensionMode, string> = {
            connected:    '$(circle-filled) Flint',
            reconnecting: '$(sync~spin) Flint',
            standalone:   '$(circle-outline) Flint (standalone)',
            degraded:     '$(warning) Flint (version mismatch)',
        };
        this.item.text = labels[mode];
    }

    setCLI(version: string, source: string) {
        this.item.tooltip = `Flint ${version} (${source})`;
    }

    setCLINotFound() {
        this.item.tooltip = 'Flint CLI not installed. Click for instructions.';
    }

    dispose() { this.item.dispose(); }
}
```

### Sidebar provider skeleton

```typescript
// extension/src/sidebar/provider.ts
import * as vscode from 'vscode';
import { Observation } from '../ipc/messages';
import { ModeManager } from '../mode';

export class SidebarProvider implements vscode.WebviewViewProvider {
    private view?: vscode.WebviewView;
    private fadeTimer?: NodeJS.Timeout;

    constructor(
        private context: vscode.ExtensionContext,
        private modeManager: ModeManager
    ) {}

    resolveWebviewView(view: vscode.WebviewView) {
        this.view = view;
        view.webview.options = { enableScripts: true };
        view.webview.html = this.getQuietStateHtml();

        view.webview.onDidReceiveMessage((msg) => {
            if (msg.type === 'dismiss') {
                this.ipc?.dismiss(msg.observationId);
                this.fadeToQuiet();
            }
            if (msg.type === 'copy') {
                vscode.env.clipboard.writeText(msg.text);
                this.fadeToQuiet();
            }
            if (msg.type === 'send') {
                this.sendToAgent(msg.prompt);
                this.fadeToQuiet();
            }
            if (msg.type === 'followUp') {
                this.streamFollowUp(msg.observationId, msg.observationText);
            }
            if (msg.type === 'close') {
                this.fadeToQuiet();
            }
        });
    }

    showObservation(obs: Observation) {
        if (!this.view) return;
        if (!this.modeManager.isPassiveIntelligenceAvailable()) return;
        clearTimeout(this.fadeTimer);
        this.view.webview.html = this.getObservationHtml(obs);
        const fadeMs = obs.kind === 'win' ? 5000 : 8000;
        this.fadeTimer = setTimeout(() => this.fadeToQuiet(), fadeMs);
    }

    private async streamFollowUp(observationId: string, observationText: string) {
        if (!this.view) return;
        clearTimeout(this.fadeTimer);
        this.view.webview.postMessage({ type: 'followUpLoading' });

        const response = await fetch('https://api.anthropic.com/v1/messages', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                model: 'claude-sonnet-4-20250514',
                max_tokens: 200,
                system: 'You are Flint, a senior developer. Give one focused follow-up in under 150 words. Be specific, concrete, direct. Senior dev tone. No preamble.',
                messages: [{ role: 'user', content: `Observation: ${observationText}\n\nMore detail please.` }],
                stream: true
            })
        });

        const reader = response.body?.getReader();
        if (!reader) return;
        let done = false;
        while (!done) {
            const { value, done: d } = await reader.read();
            done = d;
            if (value) {
                this.view.webview.postMessage({
                    type: 'followUpChunk',
                    text: new TextDecoder().decode(value)
                });
            }
        }
        this.view.webview.postMessage({ type: 'followUpDone' });
    }

    private fadeToQuiet() {
        if (this.view) this.view.webview.html = this.getQuietStateHtml();
    }

    private sendToAgent(prompt: string) {
        vscode.commands.executeCommand('cursor.chat', prompt);
    }

    private getQuietStateHtml(): string {
        return `<!DOCTYPE html><html><body style="margin:0;padding:0;">
          <div style="width:4px;height:100vh;background:var(--flint-accent);"></div>
        </body></html>`;
    }

    private getObservationHtml(obs: Observation): string {
        // Full card HTML with correct actions per card type — see PRD section 6
        return `<!-- card HTML here -->`;
    }

    private ipc: any; // injected by extension entry point
}
```



```typescript
// extension/src/sidebar/provider.ts
import * as vscode from 'vscode';
import { Observation } from '../ipc/messages';

export class SidebarProvider implements vscode.WebviewViewProvider {
    private view?: vscode.WebviewView;
    private fadeTimer?: NodeJS.Timeout;

    constructor(private context: vscode.ExtensionContext) {}

    resolveWebviewView(view: vscode.WebviewView) {
        this.view = view;
        view.webview.options = { enableScripts: true };
        view.webview.html = this.getQuietStateHtml();

    resolveWebviewView(view: vscode.WebviewView) {
        this.view = view;
        view.webview.options = { enableScripts: true };
        view.webview.html = this.getQuietStateHtml();

        view.webview.onDidReceiveMessage((msg) => {
            if (msg.type === 'dismiss') {
                this.ipc.dismiss(msg.observationId);
                this.fadeToQuiet();
            }
            if (msg.type === 'copy') {
                vscode.env.clipboard.writeText(msg.text);
                this.fadeToQuiet();
            }
            if (msg.type === 'send') {
                this.sendToAgent(msg.prompt);
                this.fadeToQuiet();
            }
            if (msg.type === 'followUp') {
                this.streamFollowUp(msg.observationId, msg.observationText);
            }
            if (msg.type === 'close') {
                this.fadeToQuiet();
            }
        });
    }

    private async streamFollowUp(observationId: string, observationText: string) {
        if (!this.view) return;
        clearTimeout(this.fadeTimer);

        // Show loading state
        this.view.webview.postMessage({ type: 'followUpLoading' });

        // Stream follow-up from Anthropic API
        const response = await fetch('https://api.anthropic.com/v1/messages', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                model: 'claude-sonnet-4-20250514',
                max_tokens: 200,
                system: 'You are Flint, a senior developer. Give one focused follow-up to this observation in under 150 words. Be specific, concrete, and direct. Senior dev tone. No preamble.',
                messages: [{ role: 'user', content: `Original observation: ${observationText}\n\nGive me more detail.` }],
                stream: true
            })
        });

        // Stream tokens to webview
        const reader = response.body?.getReader();
        if (!reader) return;
        let done = false;
        while (!done) {
            const { value, done: d } = await reader.read();
            done = d;
            if (value) {
                const text = new TextDecoder().decode(value);
                this.view.webview.postMessage({ type: 'followUpChunk', text });
            }
        }

        // Lock the card — no further follow-up possible
        this.view.webview.postMessage({ type: 'followUpDone' });
    }
    }

    showObservation(obs: Observation) {
        if (!this.view) return;
        clearTimeout(this.fadeTimer);
        this.view.webview.html = this.getObservationHtml(obs);
        this.fadeTimer = setTimeout(() => this.fadeToQuiet(), 8000);
    }

    private fadeToQuiet() {
        if (this.view) {
            this.view.webview.html = this.getQuietStateHtml();
        }
    }

    private sendToAgent(prompt: string) {
        // Agent-specific send logic
        vscode.commands.executeCommand('cursor.chat', prompt);
    }

    private getQuietStateHtml(): string {
        return `<!DOCTYPE html>
        <html><body style="margin:0;padding:0;">
          <div style="width:4px;height:100vh;background:var(--flint-accent);"></div>
        </body></html>`;
    }

    private getObservationHtml(obs: Observation): string {
        const hasPrompt = obs.agentPrompt && obs.kind === 'technical';
        return `<!DOCTYPE html>
        <html>
        <head><style>
          body { font-family: var(--vscode-font-family); padding: 12px; }
          .text { font-size: 13px; line-height: 1.6; margin-bottom: 12px; }
          .prompt { font-size: 12px; color: var(--vscode-descriptionForeground);
                    border-top: 1px solid var(--vscode-widget-border);
                    padding-top: 10px; margin-bottom: 10px; }
          .actions { display: flex; gap: 8px; }
          button { padding: 4px 12px; font-size: 12px; cursor: pointer; }
          .dismiss { position: absolute; top: 8px; right: 8px;
                     background: none; border: none; cursor: pointer; }
        </style></head>
        <body>
          <button class="dismiss" onclick="dismiss()">✕</button>
          <div class="text">${obs.text}</div>
          ${hasPrompt ? `
          <div class="prompt">${obs.agentPrompt}</div>
          <div class="actions">
            <button onclick="copy()">Copy</button>
            <button onclick="send()">Send to Cursor</button>
          </div>` : ''}
          <script>
            const vscode = acquireVsCodeApi();
            function dismiss() { vscode.postMessage({ type: 'dismiss', observationId: '${obs.id}' }); }
            function copy() { vscode.postMessage({ type: 'copy', text: \`${obs.agentPrompt}\` }); }
            function send() { vscode.postMessage({ type: 'send', prompt: \`${obs.agentPrompt}\` }); }
          </script>
        </body></html>`;
    }
}
```

---

## 8. npm Wrapper Architecture

```javascript
// wrapper/install.js — runs on npm install
const { execSync } = require('child_process');
const https = require('https');
const fs = require('fs');
const path = require('path');
const os = require('os');

const VERSION = require('./package.json').version;
const RELEASES_DIR = path.join(__dirname, 'releases');

function getPlatformBinary() {
    const platform = os.platform();
    const arch = os.arch();
    const map = {
        'darwin-arm64':  `flint-darwin-arm64`,
        'darwin-x64':    `flint-darwin-amd64`,
        'linux-x64':     `flint-linux-amd64`,
        'win32-x64':     `flint-windows-amd64.exe`,
    };
    const key = `${platform}-${arch}`;
    if (!map[key]) throw new Error(`Unsupported platform: ${key}`);
    return map[key];
}

async function download(filename) {
    const url = `https://github.com/flint-dev/flint/releases/download/v${VERSION}/${filename}`;
    const dest = path.join(RELEASES_DIR, filename);
    fs.mkdirSync(RELEASES_DIR, { recursive: true });
    // download logic
    return dest;
}

async function install() {
    const binary = getPlatformBinary();
    const dest = await download(binary);
    fs.chmodSync(dest, 0o755);
    console.log(`✓ Flint ${VERSION} installed (${binary})`);
}

install().catch(err => {
    console.error('Flint install failed:', err.message);
    process.exit(1);
});
```

```javascript
// wrapper/bin/flint.js — the shim
#!/usr/bin/env node
const { execFileSync } = require('child_process');
const path = require('path');
const os = require('os');

const binary = path.join(__dirname, '..', 'releases', getBinaryName());

try {
    execFileSync(binary, process.argv.slice(2), { stdio: 'inherit' });
} catch (e) {
    process.exit(e.status || 1);
}
```

---

## 8b. New v1 Features — Architecture Notes

### `flint scan` — onboarding scan

```go
// core/cmd/flint/scan.go
func newScanCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "scan",
        Short: "Index this codebase and establish baselines before daemon starts",
        RunE: func(cmd *cobra.Command, args []string) error {
            cfg := config.Load()
            store := store.Open(cfg.FlintDir)
            scanner := codebase.NewScanner(store)

            fmt.Println("Flint is scanning your codebase...")
            fmt.Println("This runs once. The daemon will be much more accurate after this.\n")

            result, err := scanner.Scan(cfg.ProjectPath)
            if err != nil { return err }

            fmt.Printf("  ✓ %d files indexed\n", result.FileCount)
            fmt.Printf("  ✓ %d functions analysed\n", result.FunctionCount)
            fmt.Printf("  ✓ Baselines established for %d observation categories\n", result.BaselineCount)
            fmt.Printf("  ✓ Stack detected: %s\n", strings.Join(result.Stack, ", "))
            fmt.Printf("  ✓ %d known footguns loaded for your stack\n", result.FootgunCount)
            fmt.Println("\nFlint is ready. Start the daemon with: flint start")
            return nil
        },
    }
}
```

### Footgun library — `shared/footguns.json`

```json
{
  "version": "1.0",
  "footguns": [
    {
      "id": "react-stale-closure",
      "role": "webdev",
      "framework": "react",
      "frameworkVersion": ">=16.8",
      "category": "performance",
      "severity": "high",
      "title": "Stale closure in useEffect",
      "pattern": "useEffect with dependency array missing variables used inside",
      "observation": "This useEffect references {var} but doesn't include it in the dependency array — it'll silently use the stale value from the first render.",
      "agentPrompt": "Fix the useEffect dependency array in {file} at line {line}. Add all variables referenced inside the effect to the array. Don't change the effect logic.",
      "learnMore": "https://react.dev/learn/synchronizing-with-effects#dependencies"
    },
    {
      "id": "go-goroutine-leak",
      "role": "devops",
      "framework": "go",
      "frameworkVersion": ">=1.0",
      "category": "concurrency",
      "severity": "high",
      "title": "Goroutine leak — channel never closed",
      "pattern": "goroutine started, channel passed in, no close or done signal",
      "observation": "This goroutine has no way to exit — if the channel is never closed or a done signal never sent, this leaks memory for the lifetime of the process.",
      "agentPrompt": "Add a done channel or context cancellation to the goroutine started at {file} line {line}. Ensure it can exit cleanly. Don't change the goroutine's core logic."
    }
    // ... community-contributed entries
  ]
}
```

### `flint explain` — comprehension tool

```go
// core/internal/tools/explain.go
var ExplainPrompt = `You are a senior developer explaining code to a colleague.
Given this code, explain:
1. What it does (one sentence)
2. Why it exists — what problem it solves
3. What assumptions it makes about its inputs and context
4. What would break if someone changed it carelessly

Be specific. No generic advice. Senior dev tone. Max 200 words.`
```

### `flint diff` — diff explanation tool

```go
// core/internal/tools/diff.go
var DiffPrompt = `You are a senior developer reviewing a git diff before it merges.
Given this diff, explain:
1. What actually changed (plain English, not technical)
2. What the apparent intent was
3. The top 2-3 risks this change introduces

Be specific. If you see a risk, name the exact line. Max 200 words.`
```

### Flint Watch — tripwires

```go
// core/internal/daemon/watch.go

type Tripwire struct {
    ID        string
    Pattern   string   // text pattern to watch for
    FilePath  string   // optional: specific file only
    CreatedAt time.Time
    LastFired time.Time
}

type TripwireEngine struct {
    tripwires []Tripwire
    store     *Store
}

// Tripwires bypass all time gates — they fire immediately
func (e *TripwireEngine) Check(change FileChange) *Observation {
    for _, t := range e.tripwires {
        if e.matches(change, t) {
            e.store.RecordTripwireFired(t.ID)
            return &Observation{
                Kind:       "technical",
                Category:   "tripwire",
                Confidence: 1.0, // tripwires always fire
                Text:       fmt.Sprintf("Something just touched %q — you set a watch on this.", t.Pattern),
            }
        }
    }
    return nil
}
```

CLI commands:
```bash
flint watch "payment processing"     # set tripwire on pattern
flint watch "src/auth/session.go"    # set tripwire on specific file
flint watch list                     # show all active tripwires
flint watch remove <id>              # remove a tripwire
```

### Explainability field in observation schema

```json
// addition to shared/schema.json observation type
{
  "whyINoticedThis": {
    "type": "string",
    "description": "One sentence explaining why the daemon surfaced this. Only present when confidence < 0.82."
  }
}
```

Example output:
```
That function on line 47 will throw if `user` comes back null —
happens more than you'd think on first login.

Why I noticed this: you've hit a similar null reference in this
area twice before.
```

---

## 9. Local Storage Architecture

All data lives in `~/.flint/`. SQLite via `modernc.org/sqlite` (pure Go, no CGO dependency).

```
~/.flint/
├── config.json          # role, awareness, API key, agent, human layer, auto-fix settings
├── flint.db             # main SQLite database
│   ├── observations     # all daemon observations
│   ├── error_log        # universal error log (all sources, all projects)
│   ├── error_dictionary # personal error dictionary (signature → fix mapping)
│   ├── auto_fixes       # log of every silent fix Flint has made
│   ├── calibration      # per-category confidence thresholds
│   ├── sessions         # session history for human intelligence
│   ├── human_signals    # human intelligence data points
│   ├── timegates        # last emission times per category
│   └── tool_usage       # daily tool usage counts for cross-promotion
├── errors.log           # human-readable error log (mirrors error_log table)
└── daemon.sock          # Unix socket (runtime only, deleted on stop)
```

### Schema

```sql
-- observations
CREATE TABLE observations (
    id           TEXT PRIMARY KEY,
    kind         TEXT NOT NULL,  -- technical|human|win|question|taste
    category     TEXT NOT NULL,
    confidence   REAL NOT NULL,
    text         TEXT NOT NULL,
    agent_prompt TEXT,
    why_noticed  TEXT,           -- explainability field (present when confidence < 0.82)
    file_path    TEXT,
    line_number  INTEGER,
    project      TEXT,
    session_type TEXT,           -- human|agent_assisted|vibe_coding|pair|mixed
    dismissed    INTEGER DEFAULT 0,
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- universal error log
CREATE TABLE error_log (
    id           TEXT PRIMARY KEY,
    source       TEXT NOT NULL,  -- terminal|test|build|runtime|precommit|dependency|install
    error_type   TEXT NOT NULL,  -- NullReference|BuildFailure|CVE|TestFailure|etc
    message      TEXT NOT NULL,  -- full error message
    file_path    TEXT,
    line_number  INTEGER,
    project      TEXT NOT NULL,
    signature    TEXT NOT NULL,  -- hash of error type + stack pattern (for deduplication)
    status       TEXT NOT NULL DEFAULT 'observed', -- observed|fixed|ignored|auto_fixed
    fixed_at     DATETIME,
    fix_method   TEXT,           -- manual|flint_fix|auto_fix
    fix_summary  TEXT,           -- what changed when fixed
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- personal error dictionary (built from error_log over time)
CREATE TABLE error_dictionary (
    id           TEXT PRIMARY KEY,
    signature    TEXT NOT NULL UNIQUE,
    error_type   TEXT NOT NULL,
    fix          TEXT NOT NULL,  -- what fixed it last time
    fix_count    INTEGER DEFAULT 1,
    last_seen    DATETIME,
    projects     TEXT,           -- JSON array of projects where seen
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- auto-fix log (every silent fix Flint has made)
CREATE TABLE auto_fixes (
    id           TEXT PRIMARY KEY,
    error_log_id TEXT NOT NULL,
    category     TEXT NOT NULL,  -- cve|import|format
    description  TEXT NOT NULL,  -- what Flint changed
    diff         TEXT NOT NULL,  -- the actual diff (for reference/revert)
    project      TEXT NOT NULL,
    staged       INTEGER DEFAULT 1,  -- always staged, never committed
    reverted     INTEGER DEFAULT 0,
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- calibration
CREATE TABLE calibration (
    category     TEXT PRIMARY KEY,
    threshold    REAL NOT NULL DEFAULT 0.75,
    dismissals   INTEGER DEFAULT 0,
    actions      INTEGER DEFAULT 0,
    updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- sessions
CREATE TABLE sessions (
    id                TEXT PRIMARY KEY,
    project           TEXT NOT NULL,
    session_type      TEXT,        -- dominant session type
    started_at        DATETIME NOT NULL,
    ended_at          DATETIME,
    duration_s        INTEGER,
    commit_count      INTEGER DEFAULT 0,
    revert_count      INTEGER DEFAULT 0,
    observation_count INTEGER DEFAULT 0,
    error_count       INTEGER DEFAULT 0,
    auto_fix_count    INTEGER DEFAULT 0
);

-- human signals
CREATE TABLE human_signals (
    id         TEXT PRIMARY KEY,
    kind       TEXT NOT NULL,
    value      REAL,
    context    TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- timegates
CREATE TABLE timegates (
    category   TEXT PRIMARY KEY,
    last_emit  DATETIME NOT NULL
);

-- tool usage (for cross-promotion nudge engine)
CREATE TABLE tool_usage (
    tool       TEXT NOT NULL,
    date       TEXT NOT NULL,
    count      INTEGER DEFAULT 0,
    PRIMARY KEY (tool, date)
);
```

---

## 9b. Error Log Architecture

### Error sources — what Flint captures

```go
// core/internal/errorlog/sources.go

type ErrorSource string

const (
    SourceTerminal   ErrorSource = "terminal"    // stderr from any terminal command
    SourceTest       ErrorSource = "test"         // test runner failures
    SourceBuild      ErrorSource = "build"        // compilation, bundler, type errors
    SourceRuntime    ErrorSource = "runtime"      // process crashes, panics, unhandled exceptions
    SourcePrecommit  ErrorSource = "precommit"    // hook failures, lint errors on commit
    SourceDependency ErrorSource = "dependency"   // CVE findings, health score drops
    SourceInstall    ErrorSource = "install"      // npm/pip/go mod failures
)
```

### Error logger

```go
// core/internal/errorlog/logger.go

type ErrorLogger struct {
    store   *Store
    logFile *os.File  // ~/.flint/errors.log — human readable mirror
}

func (l *ErrorLogger) Log(entry ErrorEntry) error {
    // Compute signature for deduplication and dictionary lookup
    entry.Signature = computeSignature(entry.ErrorType, entry.StackPattern)

    // Write to SQLite
    if err := l.store.InsertErrorLog(entry); err != nil {
        return err
    }

    // Mirror to human-readable log file
    l.writeToLogFile(entry)

    // Check if this matches an existing dictionary entry
    if fix := l.store.LookupDictionary(entry.Signature); fix != nil {
        // Daemon will surface this as a card
        l.notifyDaemon(entry, fix)
    }

    // Check if auto-fix is enabled for this category
    if l.shouldAutoFix(entry) {
        l.scheduleAutoFix(entry)
    }

    return nil
}

func (l *ErrorLogger) writeToLogFile(entry ErrorEntry) {
    line := fmt.Sprintf("[%s] [%s] [%s] %s",
        entry.CreatedAt.Format("2006-01-02 15:04:05"),
        entry.Source,
        entry.ErrorType,
        entry.Message,
    )
    if entry.FilePath != "" {
        line += fmt.Sprintf(" (%s:%d)", entry.FilePath, entry.LineNumber)
    }
    line += fmt.Sprintf(" [%s]\n", entry.Status)
    l.logFile.WriteString(line)
}
```

### Auto-fix engine

```go
// core/internal/errorlog/autofix.go

type AutoFixEngine struct {
    store   *Store
    ai      *AIClient
    config  *Config
    git     *GitClient
}

// Categories eligible for auto-fix (user-configurable)
type AutoFixCategory string

const (
    AutoFixCVE     AutoFixCategory = "cve"     // dependency vulnerability updates
    AutoFixImport  AutoFixCategory = "import"  // unambiguous missing imports
    AutoFixFormat  AutoFixCategory = "format"  // formatter failures (if project has one)
)

func (e *AutoFixEngine) Fix(entry ErrorEntry) error {
    if !e.config.AutoFixEnabled(AutoFixCategory(entry.Category)) {
        return ErrAutoFixNotEnabled
    }

    // Generate the fix
    fix, err := e.generateFix(entry)
    if err != nil { return err }

    // Apply it
    if err := e.applyFix(fix); err != nil { return err }

    // Stage it — NEVER commit automatically
    if err := e.git.Stage(fix.ChangedFiles); err != nil { return err }

    // Log what was done
    e.store.InsertAutoFix(AutoFixRecord{
        ErrorLogID:  entry.ID,
        Category:    entry.Category,
        Description: fix.Description,
        Diff:        fix.Diff,
        Project:     entry.Project,
        Staged:      true,
    })

    // Update error log status
    e.store.UpdateErrorStatus(entry.ID, "auto_fixed", fix.Description)

    // Notify developer via daemon → sidebar card
    e.notifyAutoFix(entry, fix)

    return nil
}

// notifyAutoFix tells the developer what was done
// Always appears as a card — developer must see every auto-fix
func (e *AutoFixEngine) notifyAutoFix(entry ErrorEntry, fix Fix) {
    obs := Observation{
        Kind:       "technical",
        Category:   "auto_fix",
        Confidence: 1.0,
        Text: fmt.Sprintf(
            "Silently fixed: %s\nChange is staged — review with `git diff --staged` before committing.",
            fix.Description,
        ),
        // No agent prompt on auto-fix cards — Flint already did the work
    }
    e.daemon.Broadcast(obs)
}
```

### Error log commands

```go
// core/cmd/flint/errors.go

// flint errors                  — recent error log (last 7 days)
// flint errors --all            — full history
// flint errors --project        — current project only
// flint errors --type cve       — filter by error type
// flint errors --status unfixed — unresolved errors only
// flint errors --clear          — delete log
// flint fix <error-id>          — fix a specific logged error
// flint fix --auto cve          — enable auto-fix for CVE category
// flint fix --auto import       — enable auto-fix for import errors
// flint fix --auto format       — enable auto-fix for formatting
// flint fix --auto off          — disable all auto-fix
// flint fix --log               — show auto-fix history
```

### Human-readable log format (`~/.flint/errors.log`)

```
[2026-05-28 09:41:23] [terminal]    [NullReference]  Cannot read property 'getId' of null (auth/session.js:47) [observed]
[2026-05-28 10:15:07] [test]        [TestFailure]     payments_test.go:23: expected 200 got 500 [observed]
[2026-05-28 11:30:44] [dependency]  [CVE]             lodash@4.17.20 CVE-2021-23337 [auto_fixed → 4.17.21]
[2026-05-28 14:22:11] [build]       [TypeError]       Type 'string' not assignable to type 'number' (utils/format.ts:12) [fixed]
[2026-05-28 15:08:33] [runtime]     [Panic]           runtime error: index out of range [0] (queue/processor.go:89) [observed]
```

Clean. Timestamped. Human readable. Always growing. Always local.

---

## 10. AI Integration Architecture

All AI calls go through a single client. Streaming is used everywhere — no waiting for full responses.

```go
// core/internal/ai/client.go
package ai

import (
    "bufio"
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

const APIBase = "https://api.anthropic.com/v1/messages"
const Model   = "claude-sonnet-4-20250514"

type Client struct {
    apiKey string
}

func NewClient(apiKey string) *Client {
    return &Client{apiKey: apiKey}
}

func (c *Client) Stream(systemPrompt, userInput string, out io.Writer) error {
    body := map[string]interface{}{
        "model":      Model,
        "max_tokens": 1024,
        "stream":     true,
        "system":     systemPrompt,
        "messages": []map[string]string{
            {"role": "user", "content": userInput},
        },
    }

    data, _ := json.Marshal(body)
    req, _ := http.NewRequest("POST", APIBase, bytes.NewReader(data))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("x-api-key", c.apiKey)
    req.Header.Set("anthropic-version", "2023-06-01")

    resp, err := http.DefaultClient.Do(req)
    if err != nil { return err }
    defer resp.Body.Close()

    scanner := bufio.NewScanner(resp.Body)
    for scanner.Scan() {
        line := scanner.Text()
        if len(line) < 6 || line[:6] != "data: " { continue }
        var event map[string]interface{}
        if err := json.Unmarshal([]byte(line[6:]), &event); err != nil { continue }
        if event["type"] == "content_block_delta" {
            if delta, ok := event["delta"].(map[string]interface{}); ok {
                if text, ok := delta["text"].(string); ok {
                    fmt.Fprint(out, text)
                }
            }
        }
    }
    return scanner.Err()
}
```

---

## 11. Build & Release Pipeline

### Makefile

```makefile
.PHONY: build dev test release clean

# Build everything
build:
	@echo "Generating schema types..."
	node scripts/generate-schema.js
	@echo "Building Go binaries..."
	cd core && go build ./cmd/flint/... ./cmd/daemon/...
	@echo "Building TS extension..."
	cd extension && npm run compile
	@echo "✓ Build complete"

# Development mode
dev:
	@echo "Starting daemon in dev mode..."
	cd core && go run ./cmd/daemon & \
	cd extension && npm run watch

# Run all tests
test:
	cd core && go test ./...
	cd extension && npm test

# Cross-compile for all platforms + publish
release:
	node scripts/generate-schema.js
	bash scripts/cross-compile.sh
	bash scripts/publish.sh

# Clean build artifacts
clean:
	rm -rf core/bin extension/out wrapper/releases
```

### Cross-compile script

```bash
#!/bin/bash
# scripts/cross-compile.sh

VERSION=$(cat wrapper/package.json | jq -r .version)
OUTPUT="wrapper/releases"
mkdir -p $OUTPUT

echo "Cross-compiling Flint v$VERSION..."

cd core

GOOS=darwin  GOARCH=arm64 go build -ldflags="-s -w" -o ../$OUTPUT/flint-darwin-arm64   ./cmd/flint
GOOS=darwin  GOARCH=amd64 go build -ldflags="-s -w" -o ../$OUTPUT/flint-darwin-amd64   ./cmd/flint
GOOS=linux   GOARCH=amd64 go build -ldflags="-s -w" -o ../$OUTPUT/flint-linux-amd64    ./cmd/flint
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ../$OUTPUT/flint-windows-amd64.exe ./cmd/flint

echo "✓ Binaries built:"
ls -lh ../$OUTPUT/
```

### CI pipeline (GitHub Actions)

```yaml
# .github/workflows/ci.yml
name: CI

on: [push, pull_request]

jobs:
  test-go:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - run: cd core && go test ./...
      - run: cd core && go vet ./...

  test-ts:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: '20' }
      - run: cd extension && npm ci && npm test

  schema-sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
      - run: node scripts/generate-schema.js
      - run: git diff --exit-code  # fail if generated files are out of sync
```

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags: ['v*']

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - uses: actions/setup-node@v4
        with: { node-version: '20' }
      - run: bash scripts/cross-compile.sh
      - name: Upload binaries to GitHub Release
        uses: softprops/action-gh-release@v2
        with:
          files: wrapper/releases/*
      - name: Publish npm wrapper
        run: cd wrapper && npm publish
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
      - name: Publish VS Code extension
        run: cd extension && npx vsce publish
        env:
          VSCE_PAT: ${{ secrets.VSCE_PAT }}
```

---

## 12. Version-by-Version Architecture Evolution

### v1 — Foundation
**What's built:** Full monorepo. Go daemon + CLI. TypeScript extension (VS Code + Cursor + Windsurf). Unix socket IPC. SQLite store. npm wrapper with cross-compiled binaries. Anthropic API integration. All 24 tools. Technical + human intelligence engines. Three awareness levels.

**Key architectural decisions locked in v1:**
- Monorepo structure — never changes
- IPC schema — versioned, backward compatible
- SQLite as local store — never replaced
- Go for runtime — never replaced
- One binary per platform — always

---

### v2 — Workflow Intelligence + Stakeholder Layer

**New components:**
```
core/internal/
├── workflow/
│   ├── habits.go         # habit pattern detection
│   ├── brief.go          # daily brief generator
│   └── personalise.go    # tool reordering logic
├── notify/
│   ├── slack.go          # Slack webhook integration
│   ├── email.go          # SMTP / SendGrid integration
│   └── telegram.go       # Telegram bot API
```

**New editor targets:**
```
plugins/
├── jetbrains/            # Kotlin — IntelliJ plugin
│   └── src/main/kotlin/
└── neovim/               # Lua
    └── lua/flint/
```

**New web surface:**
```
web/                      # Next.js — stakeholder dashboard
├── app/
├── components/
└── lib/
    └── flint-api/        # reads from shared local store via REST bridge
```

**Architecture addition — REST bridge:**
A local HTTP server (Go) that the web dashboard talks to. Serves read-only data from the local SQLite store. Never exposes data outside localhost.

```go
// core/internal/bridge/server.go
// Serves: GET /status, GET /observations, GET /human, GET /update
// Listens on: localhost:7432
// Auth: local token in ~/.flint/config.json
```

---

### v3 — Codebase Intelligence

**New components:**
```
core/internal/
├── codebase/
│   ├── index.go          # codebase indexer (AST parsing)
│   ├── graph.go          # dependency graph builder
│   ├── memory.go         # cross-project memory store
│   └── docs.go           # real-time doc generation
├── security/
│   └── patterns.go       # systemic security pattern detection
```

**Architecture addition — codebase index:**
A local vector-like index of the codebase, built incrementally. Stored in `~/.flint/index/`. Uses tree-sitter for AST parsing across languages. Enables deep context for observations without sending full codebases to the API.

```
~/.flint/
├── index/
│   ├── {project-hash}/
│   │   ├── ast/          # parsed AST snapshots
│   │   ├── graph.json    # dependency graph
│   │   └── meta.json     # index metadata
```

**Context window strategy for large codebases:**
Never send the full codebase. Build a relevance graph — when Flint needs context for an observation, it walks the dependency graph from the changed file and sends only the relevant subgraph. Max 50K tokens per API call.

**New notification channels:**
```
core/internal/notify/
├── whatsapp.go           # WhatsApp Business API
├── discord.go            # Discord webhook
└── notion.go             # Notion API integration
```

---

### v4 — Team Intelligence

**New components:**
```
core/internal/
├── team/
│   ├── sync.go           # encrypted peer-to-peer sync between team members
│   ├── aggregate.go      # team pattern aggregation (never individual)
│   ├── knowledge.go      # knowledge distribution tracking
│   └── runbooks.go       # shared runbook store
```

**Architecture addition — team sync:**
Encrypted P2P sync using public key cryptography. No central server. Team members' Flint instances sync observations and patterns directly. Individual data is never shared — only aggregated patterns.

```
~/.flint/
├── team/
│   ├── keypair.json      # local public/private key
│   ├── peers.json        # known team member public keys
│   └── shared/           # encrypted shared observations
```

**Pricing and auth layer (v4):**
```
core/internal/
└── auth/
    ├── license.go        # license key validation (local, offline-capable)
    └── team.go           # team seat management
```

---

## 13. Security Architecture

**API key storage:**
`~/.flint/config.json` is readable only by the current user (`chmod 600`). On macOS, the API key is optionally stored in the system keychain.

**No network calls except:**
1. Anthropic API — only when a tool is explicitly invoked or the daemon decides to emit an observation
2. CVE database — periodic fetch of CVE feed, cached locally in SQLite, no individual query sent
3. Binary downloads — only during `npm install`, from GitHub Releases over HTTPS

**Data that never leaves the machine:**
- All SQLite store contents
- All observation history
- All human intelligence signals
- All calibration data
- The codebase index (v3)
- Team sync data (encrypted, v4)

**IPC security:**
The Unix socket is created with `chmod 700` permissions. Only processes running as the current user can connect. On Windows, named pipes use the current user's SID.

---

## 14. Performance Targets

| Component | Target | How |
|---|---|---|
| Daemon idle CPU | < 1% | 30s evaluation tick, fsnotify (not polling) |
| Daemon idle RAM | < 30MB | Minimal Go runtime, SQLite shared cache |
| CLI startup | < 100ms | Single Go binary, no runtime startup |
| Observation emit | < 3s | Streaming API response |
| Extension load | < 500ms | Lazy activation, minimal dependencies |
| SQLite queries | < 10ms | Indexed queries, WAL mode |

---

## 15. Testing Architecture

```
core/
└── internal/
    ├── daemon/
    │   └── daemon_test.go      # daemon lifecycle, evaluation loop
    ├── technical/
    │   └── technical_test.go   # observation detection accuracy
    ├── human/
    │   └── human_test.go       # human signal detection
    ├── timegate/
    │   └── timegate_test.go    # time-gating rules
    └── ipc/
        └── ipc_test.go         # message passing

extension/
└── src/
    └── test/
        ├── sidebar.test.ts     # card rendering
        └── ipc.test.ts         # client connection + message handling
```

**The hardest thing to test:** The daemon's "when to speak" threshold. Integration tests simulate coding sessions (file changes + time passing) and assert that observations fire at the right moments and not others. These are the most important tests in the codebase.

---
