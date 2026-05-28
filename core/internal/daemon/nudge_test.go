package daemon

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"flint/core/internal/config"
	"flint/core/internal/store"
)

func tempNudgeEngine(t *testing.T, observe ObserveFn) *NudgeEngine {
	t.Helper()
	db, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	cfg := &config.Config{Awareness: config.Flame}
	return NewNudgeEngine(cfg, db, nil, t.TempDir(), observe)
}

// setLocked manipulates NudgeEngine internal state for testing.
// Tests run in the same package so they can access unexported fields directly.
func setLocked(n *NudgeEngine, fn func()) {
	n.mu.Lock()
	fn()
	n.mu.Unlock()
}

// ---------------------------------------------------------------------------
// Per-session cap
// ---------------------------------------------------------------------------

func TestNudge_PerSessionCap(t *testing.T) {
	var fired int32
	n := tempNudgeEngine(t, func(_ context.Context, _, _ string) {
		atomic.AddInt32(&fired, 1)
	})

	setLocked(n, func() {
		n.obsThisSess = humanMaxPerSess // already at cap
		n.lastChange = time.Now().Add(-humanStillness - time.Second)
	})

	n.maybeObserve(context.Background(), "/tmp/main.go")

	if atomic.LoadInt32(&fired) != 0 {
		t.Error("per-session cap should have blocked the observation")
	}
}

func TestNudge_CapResetAfterNewSession(t *testing.T) {
	var fired int32
	n := tempNudgeEngine(t, func(_ context.Context, _, _ string) {
		atomic.AddInt32(&fired, 1)
	})

	setLocked(n, func() {
		n.obsThisSess = humanMaxPerSess
	})

	n.ResetSession()

	setLocked(n, func() {
		n.lastChange = time.Now().Add(-humanStillness - time.Second)
	})

	n.maybeObserve(context.Background(), "/tmp/main.go")

	if atomic.LoadInt32(&fired) != 1 {
		t.Errorf("expected 1 observation after session reset, got %d", atomic.LoadInt32(&fired))
	}
}

// ---------------------------------------------------------------------------
// Gap gate
// ---------------------------------------------------------------------------

func TestNudge_GapGate_Blocks(t *testing.T) {
	var fired int32
	n := tempNudgeEngine(t, func(_ context.Context, _, _ string) {
		atomic.AddInt32(&fired, 1)
	})

	setLocked(n, func() {
		n.lastObs = time.Now().Add(-1 * time.Minute) // only 1 min ago, gap is 20 min
		n.lastChange = time.Now().Add(-humanStillness - time.Second)
	})

	n.maybeObserve(context.Background(), "/tmp/main.go")

	if atomic.LoadInt32(&fired) != 0 {
		t.Error("gap gate should have blocked the observation (1min < 20min gap)")
	}
}

func TestNudge_GapGate_Passes(t *testing.T) {
	var fired int32
	n := tempNudgeEngine(t, func(_ context.Context, _, _ string) {
		atomic.AddInt32(&fired, 1)
	})

	setLocked(n, func() {
		n.lastObs = time.Now().Add(-25 * time.Minute) // 25 min ago, past 20min gap
		n.lastChange = time.Now().Add(-humanStillness - time.Second)
	})

	n.maybeObserve(context.Background(), "/tmp/main.go")

	if atomic.LoadInt32(&fired) != 1 {
		t.Errorf("expected 1 observation (gap passed), got %d", atomic.LoadInt32(&fired))
	}
}

// ---------------------------------------------------------------------------
// Session boundary detection
// ---------------------------------------------------------------------------

func TestNudge_SessionBoundaryReset(t *testing.T) {
	n := tempNudgeEngine(t, nil)

	// Simulate mid-session state
	setLocked(n, func() {
		n.obsThisSess = 3
		n.lastChange = time.Now().Add(-(sessionGap + time.Minute)) // 31 min ago
	})

	// A new file event should detect the gap and reset the session
	n.OnFileChange(context.Background(), FileEvent{
		Path: "/tmp/main.go",
		Time: time.Now(),
	})

	n.mu.Lock()
	count := n.obsThisSess
	n.mu.Unlock()

	if count != 0 {
		t.Errorf("session boundary: obsThisSess should be 0 after reset, got %d", count)
	}
}

func TestNudge_NoResetWithinSession(t *testing.T) {
	n := tempNudgeEngine(t, nil)

	setLocked(n, func() {
		n.obsThisSess = 3
		n.lastChange = time.Now().Add(-5 * time.Minute) // only 5 min ago, no gap
	})

	n.OnFileChange(context.Background(), FileEvent{
		Path: "/tmp/main.go",
		Time: time.Now(),
	})

	n.mu.Lock()
	count := n.obsThisSess
	n.mu.Unlock()

	if count != 3 {
		t.Errorf("within-session: obsThisSess should stay 3, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// Stillness selection by session type
// ---------------------------------------------------------------------------

func TestNudge_StillnessBySessionType(t *testing.T) {
	n := tempNudgeEngine(t, nil)

	n.SetSessionType("vibeCoding")
	n.mu.Lock()
	d := n.stillnessLocked()
	n.mu.Unlock()
	if d != vibeStillness {
		t.Errorf("vibe stillness: want %v, got %v", vibeStillness, d)
	}

	n.SetSessionType("human")
	n.mu.Lock()
	d = n.stillnessLocked()
	n.mu.Unlock()
	if d != humanStillness {
		t.Errorf("human stillness: want %v, got %v", humanStillness, d)
	}
}
