package daemon

import (
	"context"
	"log"
	"sync"
	"time"

	"flint/core/internal/config"
	"flint/core/internal/ipc"
	"flint/core/internal/store"
)

// Time-gate constants per session type (from MVP.md and DETECTION.md).
const (
	// Human session: 90s stillness, 20min gap, 5 obs/session max
	humanStillness  = 90 * time.Second
	humanGap        = 20 * time.Minute
	humanMaxPerSess = 5

	// Vibe coding: 10s stillness, 10min gap, 8 obs/session max
	vibeStillness  = 10 * time.Second
	vibeGap        = 10 * time.Minute
	vibeMaxPerSess = 8
)

// ObserveFn is called by the nudge engine when all time gates pass.
// filePath is the file that triggered the gate; sessionType is the current classification.
type ObserveFn func(ctx context.Context, filePath, sessionType string)

// NudgeEngine applies time-gating to potential observations.
// It ensures Flint speaks up at the right moment — not constantly.
type NudgeEngine struct {
	cfg     *config.Config
	db      *store.DB
	server  *ipc.Server
	observe ObserveFn // called when gates pass; nil means gate-only mode

	mu          sync.Mutex
	lastChange  time.Time
	lastObs     time.Time
	obsThisSess int
	sessionType string

	// debouncer: a single timer, reset on every file event.
	// Replaces the previous per-event goroutine approach which leaked one
	// goroutine per save under active coding.
	timer       *time.Timer
	pendingPath string
}

// NewNudgeEngine creates a NudgeEngine.
func NewNudgeEngine(cfg *config.Config, db *store.DB, server *ipc.Server, observe ObserveFn) *NudgeEngine {
	return &NudgeEngine{
		cfg:         cfg,
		db:          db,
		server:      server,
		observe:     observe,
		sessionType: "human",
	}
}

// OnFileChange is called for every file event. It resets the debounce timer so
// the observation check fires only after a period of stillness — no matter how
// many saves arrive in between.
func (n *NudgeEngine) OnFileChange(ctx context.Context, event FileEvent) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.lastChange = event.Time
	n.pendingPath = event.Path

	// Stop the existing timer if it hasn't fired yet.
	if n.timer != nil {
		n.timer.Stop()
		// Drain the channel in case the timer already fired and sent to C
		// but the AfterFunc callback hasn't run yet. AfterFunc timers don't
		// expose C, so we just Stop — no drain needed.
	}

	stillness := n.stillnessLocked()
	// AfterFunc fires exactly once in its own goroutine — never accumulates.
	n.timer = time.AfterFunc(stillness, func() {
		n.maybeObserve(ctx, event.Path)
	})
}

// maybeObserve fires an observation if all time gates pass.
// Called from the AfterFunc goroutine — does not hold the lock on entry.
func (n *NudgeEngine) maybeObserve(ctx context.Context, triggerPath string) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	n.mu.Lock()

	now := time.Now()

	// Stillness gate: another change arrived after this timer was set
	if now.Sub(n.lastChange) < n.stillnessLocked() {
		n.mu.Unlock()
		return
	}

	// Gap gate: not enough time since the last observation
	if !n.lastObs.IsZero() && now.Sub(n.lastObs) < n.gapLocked() {
		n.mu.Unlock()
		return
	}

	// Per-session cap
	if n.obsThisSess >= n.maxLocked() {
		n.mu.Unlock()
		return
	}

	n.lastObs = now
	n.obsThisSess++
	sessionType := n.sessionType
	observeFn := n.observe

	// Release before calling observe — it may block on an AI call
	n.mu.Unlock()

	log.Printf("nudge: gate passed for %s (session: %s)", triggerPath, sessionType)

	if observeFn != nil {
		observeFn(ctx, triggerPath, sessionType)
	}
}

// SetSessionType updates the detected session type (called by the classifier).
func (n *NudgeEngine) SetSessionType(t string) {
	n.mu.Lock()
	n.sessionType = t
	n.mu.Unlock()
}

// ResetSession resets per-session counters (called when a new session starts).
func (n *NudgeEngine) ResetSession() {
	n.mu.Lock()
	n.obsThisSess = 0
	n.mu.Unlock()
}

// All three helpers below are called with n.mu already held.
func (n *NudgeEngine) stillnessLocked() time.Duration {
	if n.sessionType == "vibeCoding" {
		return vibeStillness
	}
	return humanStillness
}

func (n *NudgeEngine) gapLocked() time.Duration {
	if n.sessionType == "vibeCoding" {
		return vibeGap
	}
	return humanGap
}

func (n *NudgeEngine) maxLocked() int {
	if n.sessionType == "vibeCoding" {
		return vibeMaxPerSess
	}
	return humanMaxPerSess
}
