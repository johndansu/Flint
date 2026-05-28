package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"flint/core/internal/ai"
	"flint/core/internal/config"
	"flint/core/internal/errorlog"
	"flint/core/internal/ipc"
	"flint/core/internal/store"
	schema "flint/core/pkg/schema"
)

// Daemon is the background intelligence process.
type Daemon struct {
	cfg    *config.Config
	db     *store.DB
	server *ipc.Server
	el     *errorlog.Logger

	watcher    *Watcher
	nudge      *NudgeEngine
	pipeline   *Pipeline
	classifier *Classifier

	workspaceRoot string

	// Rolling signal accumulators for the classifier
	sigMu           sync.Mutex
	recentSaves     []time.Time     // timestamps of recent file saves
	burstLines      int             // lines changed in the current burst
	filesInBurst    map[string]bool
	hasAgentComment bool            // any file in the burst had an agent fingerprint

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New creates a new Daemon. sharedDir is the path to shared/ (for footguns.json).
// workspaceRoot is used to scope session records in the store.
func New(cfg *config.Config, db *store.DB, socketPath, sharedDir, workspaceRoot string) (*Daemon, error) {
	el, err := errorlog.New(db, config.FlintDir(), cfg.Project)
	if err != nil {
		return nil, fmt.Errorf("daemon: error log: %w", err)
	}

	d := &Daemon{cfg: cfg, db: db, el: el, workspaceRoot: workspaceRoot}
	d.server = ipc.NewServer(socketPath, d.handleMessage)

	// Build observation pipeline — AI client is nil if no key (footgun-only mode)
	var aiClient *ai.Client
	if cfg.APIKey != "" {
		aiClient = ai.New(cfg.APIKey, cfg.Model)
	}
	d.pipeline = NewPipeline(db, d.server, aiClient)

	// Load footgun patterns — non-fatal if shared dir isn't found yet
	if err := LoadFootguns(sharedDir); err != nil {
		log.Printf("footguns: %v (observations will use AI-only path)", err)
	}

	// Classifier — load persisted behavior baseline if available
	d.classifier = NewClassifier(loadBehaviorBaseline(db))
	d.filesInBurst = make(map[string]bool)

	// Wire nudge engine → observation pipeline
	d.nudge = NewNudgeEngine(cfg, db, d.server, workspaceRoot, func(ctx context.Context, filePath, sessionType string) {
		d.pipeline.Run(ctx, filePath, sessionType)
	})

	return d, nil
}

// Run starts the daemon and blocks until ctx is cancelled.
func (d *Daemon) Run(ctx context.Context, workspaceRoot string) error {
	ctx, d.cancel = context.WithCancel(ctx)

	// Start file watcher
	watcher, err := NewWatcher(workspaceRoot)
	if err != nil {
		return fmt.Errorf("daemon: watcher: %w", err)
	}
	d.watcher = watcher

	// IPC server
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		if err := d.server.Listen(); err != nil {
			log.Printf("ipc server: %v", err)
		}
	}()

	// File event loop
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.eventLoop(ctx)
	}()

	log.Printf("flintd: watching %s (awareness: %s)", workspaceRoot, d.cfg.Awareness)

	<-ctx.Done()
	d.stop()
	return nil
}

// Stop signals the daemon to shut down.
func (d *Daemon) Stop() {
	if d.cancel != nil {
		d.cancel()
	}
}

func (d *Daemon) stop() {
	if d.watcher != nil {
		d.watcher.Close()
	}
	d.server.Close()
	d.el.Close()
	d.nudge.Shutdown()
	d.wg.Wait()
}

// ---------------------------------------------------------------------------
// Event loop
// ---------------------------------------------------------------------------

func (d *Daemon) eventLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-d.watcher.Events():
			if !ok {
				return
			}
			d.onFileEvent(ctx, event)
		case err, ok := <-d.watcher.Errors():
			if !ok {
				return
			}
			_ = d.el.BuildError("file watcher: " + err.Error())
		}
	}
}

func (d *Daemon) onFileEvent(ctx context.Context, event FileEvent) {
	// Check tripwires first — they bypass time gates
	tws, err := d.db.ActiveTripwires()
	if err == nil {
		for _, tw := range tws {
			if matchesTripwire(event.Path, tw.Pattern) {
				d.fireTripwire(tw, event.Path)
			}
		}
	}

	// Check for agent-generated content (file is already in page cache after save)
	if content, err := os.ReadFile(event.Path); err == nil {
		if HasAgentFingerprint(string(content)) {
			d.sigMu.Lock()
			d.hasAgentComment = true
			d.sigMu.Unlock()
		}
	}

	// Update rolling signal accumulators for the classifier
	d.updateSignals(event)

	// Pass to nudge engine for time-gated analysis
	d.nudge.OnFileChange(ctx, event)
}

// updateSignals maintains the rolling window used by the session classifier.
// Runs on every file event and re-classifies every 10 saves.
func (d *Daemon) updateSignals(event FileEvent) {
	d.sigMu.Lock()
	defer d.sigMu.Unlock()

	now := event.Time

	// Keep only the last 2 minutes of saves
	cutoff := now.Add(-2 * time.Minute)
	fresh := d.recentSaves[:0]
	for _, t := range d.recentSaves {
		if t.After(cutoff) {
			fresh = append(fresh, t)
		}
	}
	d.recentSaves = append(fresh, now)
	d.filesInBurst[event.Path] = true

	// Re-classify every 10 saves
	if len(d.recentSaves)%10 != 0 {
		return
	}

	// Build a signal snapshot from what we know without reading files
	var writeIntervalMs int64 = 800
	if n := len(d.recentSaves); n >= 2 {
		writeIntervalMs = d.recentSaves[n-1].Sub(d.recentSaves[n-2]).Milliseconds()
	}

	savesPerMin := float64(len(d.recentSaves)) / 2.0 // saves over 2-minute window

	sig := Signal{
		WriteIntervalMs: writeIntervalMs,
		SavesPerMinute:  savesPerMin,
		FilesChanged:    len(d.filesInBurst),
		HasLargeChunk:   d.burstLines > 50,
		BurstSize:       d.burstLines,
		HasAgentComment: d.hasAgentComment,
		AgentToolActive: detectAgentToolActive(),
		MultipleCursors: detectMultipleAuthors(d.workspaceRoot),
	}

	sessionType, _ := d.classifier.Classify(sig)
	d.nudge.SetSessionType(string(sessionType))

	// Broadcast updated session state
	since := time.Now().UTC().Format(time.RFC3339)
	d.server.BroadcastSessionState(schema.SessionState{
		State:       "deepWork",
		SessionType: string(sessionType),
		Since:       since,
	})

	// Blend this cycle's measurements into the persisted behavior baseline
	d.updateBehaviorBaseline(sig.WriteIntervalMs, sig.SavesPerMinute, sig.BurstSize)

	// Reset burst accumulators each classification cycle
	d.burstLines = 0
	d.filesInBurst = make(map[string]bool)
	d.hasAgentComment = false
}

func (d *Daemon) fireTripwire(tw store.Tripwire, filePath string) {
	tf := schema.TripwireFired{
		TripwireId: tw.ID,
		Pattern:    tw.Pattern,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}
	if filePath != "" {
		tf.FilePath = &filePath
	}
	d.server.BroadcastTripwireFired(tf)
	log.Printf("tripwire fired: %q matched %s", tw.Pattern, filePath)
}

// ---------------------------------------------------------------------------
// IPC message handling
// ---------------------------------------------------------------------------

func (d *Daemon) handleMessage(clientID string, env ipc.Envelope) {
	switch env.Type {
	case "subscribe":
		// Client just connected — send current status
		status := schema.DaemonStatus{
			Awareness:        string(d.cfg.Awareness),
			HumanLayer:       d.cfg.HumanLayer,
			ObservationCount: 0, // TODO: read from store
		}
		if d.cfg.Project != "" {
			status.Project = &d.cfg.Project
		}
		d.server.Broadcast("daemonStatus", status)

	case "dismissal":
		var dismissal schema.Dismissal
		if err := json.Unmarshal(env.Payload, &dismissal); err == nil {
			d.server.Broadcast("ack", schema.Ack{
				ObservationId: dismissal.ObservationId,
				Action:        "dismissed",
			})
		}
	}
}

// ---------------------------------------------------------------------------
// Baseline helpers
// ---------------------------------------------------------------------------

// loadBehaviorBaseline reads persisted behavior averages from the store.
// Falls back to conservative hardcoded defaults when no data exists yet.
func loadBehaviorBaseline(db *store.DB) *Baseline {
	b := &Baseline{
		AvgWriteIntervalMs: 800,
		AvgBurstSize:       3,
		AvgSavesPerMinute:  2,
	}
	if v, _ := db.GetBaseline("behavior.avg_write_interval_ms"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			b.AvgWriteIntervalMs = f
		}
	}
	if v, _ := db.GetBaseline("behavior.avg_burst_size"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			b.AvgBurstSize = f
		}
	}
	if v, _ := db.GetBaseline("behavior.avg_saves_per_minute"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			b.AvgSavesPerMinute = f
		}
	}
	if v, _ := db.GetBaseline("behavior.session_count"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			b.SessionCount = n
		}
	}
	return b
}

// updateBehaviorBaseline applies an exponential moving average (α=0.1) to
// blend the current cycle's measurements into the persisted baseline.
func (d *Daemon) updateBehaviorBaseline(writeIntervalMs int64, savesPerMin float64, burstSize int) {
	const alpha = 0.1
	b := d.classifier.baseline

	b.AvgWriteIntervalMs = alpha*float64(writeIntervalMs) + (1-alpha)*b.AvgWriteIntervalMs
	b.AvgSavesPerMinute  = alpha*savesPerMin + (1-alpha)*b.AvgSavesPerMinute
	b.AvgBurstSize       = alpha*float64(burstSize) + (1-alpha)*b.AvgBurstSize
	b.SessionCount++

	_ = d.db.SetBaseline("behavior.avg_write_interval_ms", strconv.FormatFloat(b.AvgWriteIntervalMs, 'f', 2, 64))
	_ = d.db.SetBaseline("behavior.avg_saves_per_minute",  strconv.FormatFloat(b.AvgSavesPerMinute, 'f', 2, 64))
	_ = d.db.SetBaseline("behavior.avg_burst_size",        strconv.FormatFloat(b.AvgBurstSize, 'f', 2, 64))
	_ = d.db.SetBaseline("behavior.session_count",         strconv.Itoa(b.SessionCount))
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func matchesTripwire(filePath, pattern string) bool {
	if strings.Contains(filePath, pattern) {
		return true
	}
	return fileContains(filePath, pattern)
}

func fileContains(path, pattern string) bool {
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(b), pattern)
}
