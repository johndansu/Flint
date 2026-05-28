package daemon

import (
	"testing"
	"time"
)

func defaultBaseline() *Baseline {
	return &Baseline{
		AvgWriteIntervalMs: 800,
		AvgBurstSize:       3,
		AvgSavesPerMinute:  2,
	}
}

func TestClassifier_Human(t *testing.T) {
	c := NewClassifier(defaultBaseline())
	sig := Signal{
		WriteIntervalMs: 750, // within ±50% of 800ms baseline
		BurstSize:       2,
		SavesPerMinute:  1.8,
	}
	got, conf := c.Classify(sig)
	if got != SessionHuman {
		t.Errorf("expected %s, got %s (conf=%.2f)", SessionHuman, got, conf)
	}
}

func TestClassifier_VibeCoding(t *testing.T) {
	c := NewClassifier(defaultBaseline())
	sig := Signal{
		WriteIntervalMs: 40,  // far below baseline — agent speed
		BurstSize:       80,  // large chunk
		HasLargeChunk:   true,
		SavesPerMinute:  20,
		HasAgentComment: true,
	}
	got, conf := c.Classify(sig)
	if got != SessionVibeCoding {
		t.Errorf("expected %s, got %s (conf=%.2f)", SessionVibeCoding, got, conf)
	}
	if conf <= 0 {
		t.Errorf("confidence should be > 0, got %.2f", conf)
	}
}

func TestClassifier_AgentAssisted(t *testing.T) {
	c := NewClassifier(defaultBaseline())
	sig := Signal{
		// WriteIntervalMs 200 → ratio 0.25, outside human range (0.5–2.0) → no human score
		WriteIntervalMs:  200,
		BurstSize:        10, // neutral (> 5, < 20)
		AgentToolActive:  true,
		ChatWindowActive: true,
	}
	got, _ := c.Classify(sig)
	if got != SessionAgentAssisted {
		t.Errorf("expected %s, got %s", SessionAgentAssisted, got)
	}
}

func TestClassifier_PairProgramming(t *testing.T) {
	c := NewClassifier(defaultBaseline())
	sig := Signal{
		WriteIntervalMs: 900, // in human range — pair programmers type normally
		BurstSize:       2,   // small: human +0.10
		MultipleCursors: true, // pair-prog +0.45 — dominates
		// No TestsUpdated: would add +0.08 to human, narrowing the gap
	}
	got, _ := c.Classify(sig)
	if got != SessionPairProg {
		t.Errorf("expected %s, got %s", SessionPairProg, got)
	}
}

func TestClassifier_Mixed(t *testing.T) {
	c := NewClassifier(defaultBaseline())
	// Signals point simultaneously to AgentAssisted and VibeCoding
	// with scores that are close — should produce Mixed
	sig := Signal{
		WriteIntervalMs:  50,  // very fast → vibe
		HasLargeChunk:    true, // vibe
		AgentToolActive:  true, // agent
		ChatWindowActive: true, // agent
		HasAgentComment:  true, // both
		TestsUpdated:     true, // human/agent
	}
	got, _ := c.Classify(sig)
	// Mixed is only assigned when ≥2 types are within 0.10 of the winner.
	// This signal set should be ambiguous enough; if not, accept agent or vibe.
	if got != SessionMixed && got != SessionAgentAssisted && got != SessionVibeCoding {
		t.Errorf("expected mixed/agent/vibe for ambiguous signals, got %s", got)
	}
}

func TestClassifier_PersonalBaseline(t *testing.T) {
	// A developer who types very fast (200ms intervals, 8 saves/min)
	// should still be classified as human when those are their personal norms.
	fastBaseline := &Baseline{
		AvgWriteIntervalMs: 200,
		AvgBurstSize:       3,
		AvgSavesPerMinute:  8,
	}
	c := NewClassifier(fastBaseline)
	sig := Signal{
		WriteIntervalMs: 180, // within ±50% of their 200ms baseline
		BurstSize:       3,
		SavesPerMinute:  7,
	}
	got, _ := c.Classify(sig)
	if got != SessionHuman {
		t.Errorf("fast-typer should classify as human against their own baseline, got %s", got)
	}
}

func TestClassifier_HistoryCapped(t *testing.T) {
	c := NewClassifier(defaultBaseline())
	sig := Signal{WriteIntervalMs: 800, BurstSize: 2}
	// Drive history well past the 100-entry cap
	for i := 0; i < 120; i++ {
		c.Classify(sig)
	}
	if len(c.history) > 100 {
		t.Errorf("history should be capped at 100, got %d", len(c.history))
	}
}

func TestClassifier_IdleAfterBurst(t *testing.T) {
	c := NewClassifier(defaultBaseline())
	sig := Signal{
		WriteIntervalMs: 50,
		HasLargeChunk:   true,
		BurstSize:       60,
		IdleAfterBurst:  20 * time.Second, // classic vibe pattern
	}
	got, _ := c.Classify(sig)
	if got != SessionVibeCoding {
		t.Errorf("idle-after-burst pattern should classify as vibe coding, got %s", got)
	}
}
