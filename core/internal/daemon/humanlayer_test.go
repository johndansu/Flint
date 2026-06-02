package daemon

import (
	"testing"
	"time"

	"flint/core/internal/config"
	"flint/core/internal/store"
)

func newTestHumanLayer(t *testing.T) *HumanLayer {
	t.Helper()
	db, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewHumanLayer(nil, db, t.TempDir(), config.Thresholds{})
}

// ---------------------------------------------------------------------------
// Spiral detection
// ---------------------------------------------------------------------------

func TestHumanLayer_SpiralNotFiredBelowThreshold(t *testing.T) {
	h := newTestHumanLayer(t)
	now := time.Now()

	// 9 edits — one below the default threshold of 10
	for i := 0; i < 9; i++ {
		h.OnFileEvent(FileEvent{Path: "/proj/main.go", Time: now.Add(time.Duration(i) * time.Second)})
	}

	h.mu.Lock()
	_, fired := h.spiralFired["/proj/main.go"]
	h.mu.Unlock()

	if fired {
		t.Error("spiral should not fire below threshold (9 < 10)")
	}
}

func TestHumanLayer_SpiralFiredAtThreshold(t *testing.T) {
	h := newTestHumanLayer(t)
	now := time.Now()

	for i := 0; i < 10; i++ {
		h.OnFileEvent(FileEvent{Path: "/proj/main.go", Time: now.Add(time.Duration(i) * time.Second)})
	}

	h.mu.Lock()
	firedAt, fired := h.spiralFired["/proj/main.go"]
	h.mu.Unlock()

	if !fired {
		t.Error("spiral should fire at threshold (10 edits)")
	}
	if firedAt.IsZero() {
		t.Error("spiral fired time should not be zero")
	}
}

func TestHumanLayer_SpiralCustomThreshold(t *testing.T) {
	db, _ := store.Open(t.TempDir())
	defer db.Close()
	h := NewHumanLayer(nil, db, t.TempDir(), config.Thresholds{SpiralEdits: 5})
	now := time.Now()

	for i := 0; i < 5; i++ {
		h.OnFileEvent(FileEvent{Path: "/proj/auth.go", Time: now.Add(time.Duration(i) * time.Second)})
	}

	h.mu.Lock()
	_, fired := h.spiralFired["/proj/auth.go"]
	h.mu.Unlock()

	if !fired {
		t.Error("spiral should fire at custom threshold of 5")
	}
}

func TestHumanLayer_SpiralEditsOutsideWindowNotCounted(t *testing.T) {
	h := newTestHumanLayer(t)

	// 5 edits 40 minutes ago — outside the 30-min window
	old := time.Now().Add(-40 * time.Minute)
	for i := 0; i < 5; i++ {
		h.OnFileEvent(FileEvent{Path: "/proj/x.go", Time: old.Add(time.Duration(i) * time.Second)})
	}

	// 5 recent edits — total recent count is 5, under threshold
	now := time.Now()
	for i := 0; i < 5; i++ {
		h.OnFileEvent(FileEvent{Path: "/proj/x.go", Time: now.Add(time.Duration(i) * time.Second)})
	}

	h.mu.Lock()
	_, fired := h.spiralFired["/proj/x.go"]
	h.mu.Unlock()

	if fired {
		t.Error("old edits outside the window should not count toward spiral threshold")
	}
}

func TestHumanLayer_SpiralCooldownPreventsRefire(t *testing.T) {
	h := newTestHumanLayer(t)
	now := time.Now()

	// Fire the spiral
	for i := 0; i < 10; i++ {
		h.OnFileEvent(FileEvent{Path: "/proj/m.go", Time: now.Add(time.Duration(i) * time.Second)})
	}

	h.mu.Lock()
	// Simulate that fire happened 1 minute ago (within cooldown)
	h.spiralFired["/proj/m.go"] = time.Now().Add(-1 * time.Minute)
	h.mu.Unlock()

	// 10 more edits in a fresh window — should be blocked by cooldown
	later := time.Now()
	for i := 0; i < 10; i++ {
		h.OnFileEvent(FileEvent{Path: "/proj/m.go", Time: later.Add(time.Duration(i) * time.Second)})
	}

	h.mu.Lock()
	firedAt := h.spiralFired["/proj/m.go"]
	h.mu.Unlock()

	// If cooldown worked, spiralFired was NOT updated — it's still ~1 minute ago.
	// A freshly-updated timestamp would be within the last second.
	if time.Since(firedAt) < time.Second {
		t.Error("spiral cooldown should prevent re-firing: spiralFired was updated when it should have been blocked")
	}
}

// ---------------------------------------------------------------------------
// Burnout detection
// ---------------------------------------------------------------------------

func TestHumanLayer_BurnoutNotFiredBeforeThreshold(t *testing.T) {
	h := newTestHumanLayer(t)
	h.mu.Lock()
	h.sessionStart = time.Now().Add(-2 * time.Hour) // 2h < 3h threshold
	h.mu.Unlock()

	// CheckBurnout should not fire
	h.CheckBurnout()

	h.mu.Lock()
	fired := !h.lastBurnoutFired.IsZero()
	h.mu.Unlock()

	if fired {
		t.Error("burnout should not fire before 3-hour threshold")
	}
}

func TestHumanLayer_BurnoutCustomThreshold(t *testing.T) {
	db, _ := store.Open(t.TempDir())
	defer db.Close()
	h := NewHumanLayer(nil, db, t.TempDir(), config.Thresholds{BurnoutHours: 1})

	h.mu.Lock()
	h.sessionStart = time.Now().Add(-70 * time.Minute) // 70min > 1h custom threshold
	h.mu.Unlock()

	h.CheckBurnout()

	h.mu.Lock()
	fired := !h.lastBurnoutFired.IsZero()
	h.mu.Unlock()

	if !fired {
		t.Error("burnout should fire with custom 1-hour threshold after 70 minutes")
	}
}

func TestHumanLayer_BurnoutResetOnNewSession(t *testing.T) {
	h := newTestHumanLayer(t)
	h.mu.Lock()
	h.sessionStart = time.Now().Add(-4 * time.Hour)
	h.mu.Unlock()

	h.OnNewSession()

	h.mu.Lock()
	age := time.Since(h.sessionStart)
	h.mu.Unlock()

	if age > time.Minute {
		t.Errorf("OnNewSession should reset sessionStart to now, got age=%v", age)
	}
}
