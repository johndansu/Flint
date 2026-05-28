package nudge

import (
	"context"
	"fmt"
	"time"

	"flint/core/internal/ai"
	"flint/core/internal/config"
	"flint/core/internal/store"
	schema "flint/core/pkg/schema"

	"github.com/google/uuid"
)

// HumanObserver watches developer patterns and fires human-layer observations.
// It is only active when config.HumanLayer is true (opt-in, off by default).
type HumanObserver struct {
	cfg    *config.Config
	db     *store.DB
	client *ai.Client

	// Pattern accumulators
	sessionStart   time.Time
	debuggingFor   time.Duration
	lastBreak      time.Time
	errorCount     int
	sameErrorCount int
	lastErrorMsg   string
}

// NewHumanObserver creates an observer. Returns nil without error if HumanLayer is disabled.
func NewHumanObserver(cfg *config.Config, db *store.DB) *HumanObserver {
	if !cfg.HumanLayer {
		return nil
	}
	if cfg.APIKey == "" {
		return nil
	}
	return &HumanObserver{
		cfg:          cfg,
		db:           db,
		client:       ai.New(cfg.APIKey, cfg.Model),
		sessionStart: time.Now(),
		lastBreak:    time.Now(),
	}
}

// Observe is called periodically to check for human-layer signals.
// Fires at most one observation per call; returns nil if nothing warrants attention.
func (h *HumanObserver) Observe(ctx context.Context, signal HumanSignal) (*schema.Observation, error) {
	if h == nil {
		return nil, nil
	}

	h.updateAccumulators(signal)

	// Check patterns in priority order — first match wins per call
	if obs := h.checkDebuggingSpiral(); obs != nil {
		return h.enrich(ctx, obs)
	}
	if obs := h.checkBurnoutSignal(); obs != nil {
		return h.enrich(ctx, obs)
	}
	if obs := h.checkLoneWolf(signal); obs != nil {
		return h.enrich(ctx, obs)
	}

	return nil, nil
}

// HumanSignal is a snapshot of observable developer behaviour.
type HumanSignal struct {
	ActiveMinutes    int
	MinutesSinceBreak int
	ErrorsLastHour   int
	SameErrorRepeated bool
	CommitsToday     int
	PRsReviewedThisWeek int
	FilesAlwaysAlone bool // developer only ever touches certain files
}

// ---------------------------------------------------------------------------
// Pattern detectors
// ---------------------------------------------------------------------------

func (h *HumanObserver) checkDebuggingSpiral() *schema.Observation {
	// Debugging spiral: stuck on same error for > 45 minutes
	if h.sameErrorCount >= 3 && h.debuggingFor > 45*time.Minute {
		text := fmt.Sprintf(
			"You've been hitting the same error category for %s. "+
				"Sometimes a rubber duck or a 10-minute walk breaks the loop faster than another debug cycle.",
			roundDuration(h.debuggingFor),
		)
		return h.makeObs("question", "debugging-spiral", text, 0.78)
	}
	return nil
}

func (h *HumanObserver) checkBurnoutSignal() *schema.Observation {
	// Burnout signal: active for > 4 hours without a break > 10 minutes
	sessionLen := time.Since(h.sessionStart)
	timeSinceBreak := time.Since(h.lastBreak)

	if sessionLen > 4*time.Hour && timeSinceBreak > 2*time.Hour {
		text := "You've been coding for over 4 hours without a real break. " +
			"Decision quality degrades after 90 minutes of focused work. " +
			"Ten minutes now is worth an hour later."
		return h.makeObs("human", "burnout-signal", text, 0.72)
	}
	return nil
}

func (h *HumanObserver) checkLoneWolf(s HumanSignal) *schema.Observation {
	// Lone wolf: always working alone on the same files for extended period
	if s.FilesAlwaysAlone && s.PRsReviewedThisWeek == 0 && s.CommitsToday > 5 {
		text := "You're the only person who has touched this part of the codebase in a while. " +
			"A single review request broadens the knowledge base and catches blind spots. " +
			"Bus factor of one on critical code is a risk worth naming."
		return h.makeObs("human", "lone-wolf", text, 0.65)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (h *HumanObserver) updateAccumulators(s HumanSignal) {
	if s.MinutesSinceBreak < 5 {
		h.lastBreak = time.Now()
		h.debuggingFor = 0
		h.sameErrorCount = 0
	}

	if s.SameErrorRepeated {
		h.sameErrorCount++
		h.debuggingFor += time.Minute
	} else {
		h.sameErrorCount = 0
		h.debuggingFor = 0
	}

	h.errorCount = s.ErrorsLastHour
}

func (h *HumanObserver) makeObs(kind, category, text string, confidence float64) *schema.Observation {
	id := uuid.New().String()
	ts := time.Now().UTC().Format(time.RFC3339)

	var explainWhy *string
	if confidence < 0.82 {
		why := fmt.Sprintf("Confidence %.0f%% — pattern matched but not conclusive", confidence*100)
		explainWhy = &why
	}

	return &schema.Observation{
		Id:          id,
		Kind:        kind,
		Category:    category,
		Confidence:  confidence,
		Text:        text,
		ExplainWhy:  explainWhy,
		Timestamp:   ts,
		SessionType: strPtr("human"),
	}
}

// enrich optionally calls the AI to extend the observation's agentPrompt field.
// This is a best-effort call — if it fails the base observation is returned as-is.
func (h *HumanObserver) enrich(ctx context.Context, obs *schema.Observation) (*schema.Observation, error) {
	// For now, return the observation without AI enrichment.
	// Phase 9 will wire this to generate agentPrompt suggestions.
	return obs, nil
}

func strPtr(s string) *string { return &s }

func roundDuration(d time.Duration) string {
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	default:
		return fmt.Sprintf("%.1f hours", d.Hours())
	}
}
