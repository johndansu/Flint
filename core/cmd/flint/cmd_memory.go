package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"flint/core/internal/config"
	"flint/core/internal/store"
)

func cmdMemory(args []string) error {
	db, err := store.Open(config.FlintDir())
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer db.Close()

	sub := ""
	if len(args) > 0 {
		sub = args[0]
	}

	switch sub {
	case "clear":
		return memoryClear(db, args[1:])
	case "obs", "observations":
		return memoryObs(db)
	case "errors":
		return memoryErrors(db)
	case "baselines":
		return memoryBaselines(db)
	case "--export", "export":
		return memoryExport(db)
	default:
		return memorySummary(db)
	}
}

// ---------------------------------------------------------------------------
// Summary (default)
// ---------------------------------------------------------------------------

func memorySummary(db *store.DB) error {
	fmt.Printf("\nFlint memory\n%s\n", separator())

	// Observation counts by kind
	counts, err := db.ObservationKindCounts()
	if err != nil {
		return err
	}
	total := 0
	for _, n := range counts {
		total += n
	}
	fmt.Printf("  Observations: %d total\n", total)
	for _, kind := range []string{"technical", "human", "win", "taste", "question"} {
		if n := counts[kind]; n > 0 {
			fmt.Printf("    %-12s %d\n", kind, n)
		}
	}

	// Error count
	if errs, err := db.RecentErrors(1); err == nil {
		// Use a rough count via querying all baselines — just show recent errors
		_ = errs
	}
	recentErrs, err := db.RecentErrors(5)
	if err == nil {
		fmt.Printf("\n  Errors logged: %d (showing last %d)\n", len(recentErrs), len(recentErrs))
		for _, e := range recentErrs {
			msg := e.Message
			if len(msg) > 60 {
				msg = msg[:60] + "…"
			}
			fmt.Printf("    #%-4d [%s] %s\n", e.ID, e.Source, msg)
		}
	}

	fmt.Printf("\n  Sub-commands:\n")
	fmt.Printf("    flint memory obs           Recent observations\n")
	fmt.Printf("    flint memory errors        Recent errors\n")
	fmt.Printf("    flint memory baselines     Stored baselines\n")
	fmt.Printf("    flint memory export        Export all data as JSON\n")
	fmt.Printf("    flint memory clear obs     Clear all observations\n")
	fmt.Printf("    flint memory clear errors  Clear all errors\n")
	fmt.Printf("    flint memory clear all     Clear all transient data\n")
	fmt.Println()
	return nil
}

// ---------------------------------------------------------------------------
// Sub-views
// ---------------------------------------------------------------------------

func memoryObs(db *store.DB) error {
	obs, err := db.RecentObservations(20)
	if err != nil {
		return err
	}
	fmt.Printf("\nRecent observations (%d)\n%s\n", len(obs), separator())
	if len(obs) == 0 {
		fmt.Println("  (none)")
	}
	for _, o := range obs {
		ts := o.Timestamp.Format("2006-01-02 15:04")
		conf := fmt.Sprintf("%.0f%%", o.Confidence*100)
		shortID := o.ID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}
		loc := ""
		if o.FilePath != "" {
			loc = " · " + shortPath(o.FilePath)
			if o.LineNumber > 0 {
				loc += fmt.Sprintf(":%d", o.LineNumber)
			}
		}
		text := o.Text
		if len(text) > 70 {
			text = text[:70] + "…"
		}
		fmt.Printf("  %s  [%s] %s %-12s %s%s\n", shortID, ts, conf, o.Kind+"/"+o.Category, text, loc)
	}
	fmt.Println()
	fmt.Println("  Dismiss or calibrate: flint dismiss <id>")
	fmt.Println()
	return nil
}

func memoryErrors(db *store.DB) error {
	errs, err := db.RecentErrors(20)
	if err != nil {
		return err
	}
	fmt.Printf("\nRecent errors (%d)\n%s\n", len(errs), separator())
	if len(errs) == 0 {
		fmt.Println("  (none)")
	}
	for _, e := range errs {
		ts := e.Timestamp.Format("2006-01-02 15:04")
		loc := ""
		if e.FilePath != "" {
			loc = " · " + shortPath(e.FilePath)
			if e.LineNumber > 0 {
				loc += fmt.Sprintf(":%d", e.LineNumber)
			}
		}
		msg := e.Message
		if len(msg) > 80 {
			msg = msg[:80] + "…"
		}
		fmt.Printf("  #%-4d [%s] %s  %s%s\n", e.ID, ts, e.Source, msg, loc)
	}
	fmt.Println()
	return nil
}

func memoryBaselines(db *store.DB) error {
	baselines, err := db.AllBaselines()
	if err != nil {
		return err
	}
	fmt.Printf("\nBaselines (%d)\n%s\n", len(baselines), separator())
	for _, kv := range baselines {
		val := kv[1]
		if len(val) > 60 {
			val = val[:60] + "…"
		}
		fmt.Printf("  %-45s %s\n", kv[0], val)
	}
	fmt.Println()
	return nil
}

// ---------------------------------------------------------------------------
// Clear
// ---------------------------------------------------------------------------

func memoryClear(db *store.DB, args []string) error {
	target := "all"
	if len(args) > 0 {
		target = args[0]
	}

	switch target {
	case "obs", "observations":
		if err := db.ClearObservations(); err != nil {
			return err
		}
		fmt.Println("Cleared all observations.")
	case "errors":
		if err := db.ClearErrors(); err != nil {
			return err
		}
		fmt.Println("Cleared all errors.")
	case "all":
		if err := db.ClearAll(); err != nil {
			return err
		}
		fmt.Println("Cleared all observations, errors, dismissals, and auto-fixes.")
	default:
		return fmt.Errorf("unknown target %q — use: obs, errors, all", target)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Export
// ---------------------------------------------------------------------------

type memoryExportDoc struct {
	ExportedAt   string              `json:"exported_at"`
	Observations []store.Observation `json:"observations"`
	Baselines    map[string]string   `json:"baselines"`
}

func memoryExport(db *store.DB) error {
	obs, err := db.RecentObservations(0) // 0 = unlimited via the query below
	if err != nil {
		return err
	}
	// RecentObservations uses LIMIT ?, so pass a large number instead.
	obs, err = db.RecentObservations(100000)
	if err != nil {
		return err
	}

	kvs, err := db.AllBaselines()
	if err != nil {
		return err
	}
	baselines := make(map[string]string, len(kvs))
	for _, kv := range kvs {
		baselines[kv[0]] = kv[1]
	}

	doc := memoryExportDoc{
		ExportedAt:   time.Now().UTC().Format(time.RFC3339),
		Observations: obs,
		Baselines:    baselines,
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(doc)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func shortPath(p string) string {
	parts := strings.Split(p, "/")
	if len(parts) <= 2 {
		return p
	}
	return strings.Join(parts[len(parts)-2:], "/")
}
