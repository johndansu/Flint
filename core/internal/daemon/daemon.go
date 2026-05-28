package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
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

	// Rolling signal accumulators for the classifier
	sigMu       sync.Mutex
	recentSaves []time.Time // timestamps of recent file saves
	burstLines  int         // lines changed in the current burst
	filesInBurst map[string]bool

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New creates a new Daemon. sharedDir is the path to shared/ (for footguns.json).
func New(cfg *config.Config, db *store.DB, socketPath, sharedDir string) (*Daemon, error) {
	el, err := errorlog.New(db, config.FlintDir(), cfg.Project)
	if err != nil {
		return nil, fmt.Errorf("daemon: error log: %w", err)
	}

	d := &Daemon{cfg: cfg, db: db, el: el}
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

	// Classifier with default baseline (replaced by real baseline after flint scan)
	d.classifier = NewClassifier(nil)
	d.filesInBurst = make(map[string]bool)

	// Wire nudge engine → observation pipeline
	d.nudge = NewNudgeEngine(cfg, db, d.server, func(ctx context.Context, filePath, sessionType string) {
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

	// Reset burst counter each classification cycle
	d.burstLines = 0
	d.filesInBurst = make(map[string]bool)
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
