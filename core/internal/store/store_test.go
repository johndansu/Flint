package store

import (
	"testing"
	"time"
)

func tempDB(t *testing.T) *DB {
	t.Helper()
	db, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// ---------------------------------------------------------------------------
// Observations
// ---------------------------------------------------------------------------

func TestInsertAndCountObservations(t *testing.T) {
	db := tempDB(t)

	n, err := db.CountObservations()
	if err != nil || n != 0 {
		t.Fatalf("expected 0 observations, got %d (%v)", n, err)
	}

	obs := Observation{
		ID:         "obs-1",
		Kind:       "technical",
		Category:   "security",
		Confidence: 0.9,
		Text:       "You have a potential SQL injection here.",
		Timestamp:  time.Now(),
	}
	if err := db.InsertObservation(obs); err != nil {
		t.Fatalf("InsertObservation: %v", err)
	}

	n, err = db.CountObservations()
	if err != nil {
		t.Fatalf("CountObservations: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 observation, got %d", n)
	}
}

func TestInsertObservation_Idempotent(t *testing.T) {
	db := tempDB(t)
	obs := Observation{
		ID:        "obs-dup",
		Kind:      "technical",
		Category:  "perf",
		Confidence: 0.7,
		Text:      "N+1 query detected.",
		Timestamp: time.Now(),
	}
	if err := db.InsertObservation(obs); err != nil {
		t.Fatal(err)
	}
	// INSERT OR IGNORE — second insert must not error
	if err := db.InsertObservation(obs); err != nil {
		t.Errorf("duplicate insert should be ignored, got: %v", err)
	}
	n, _ := db.CountObservations()
	if n != 1 {
		t.Errorf("expected 1 observation after duplicate insert, got %d", n)
	}
}

// ---------------------------------------------------------------------------
// Sessions
// ---------------------------------------------------------------------------

func TestStartAndEndSession(t *testing.T) {
	db := tempDB(t)

	id, err := db.StartSession("human", "/projects/myapp")
	if err != nil {
		t.Fatalf("StartSession: %v", err)
	}
	if id == 0 {
		t.Fatal("StartSession returned zero ID")
	}

	if err := db.EndSession(id); err != nil {
		t.Fatalf("EndSession: %v", err)
	}

	// Ending again should be a no-op (UPDATE with no row → no error)
	if err := db.EndSession(id + 999); err != nil {
		t.Errorf("EndSession on missing ID should not error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Errors
// ---------------------------------------------------------------------------

func TestInsertAndGetError(t *testing.T) {
	db := tempDB(t)

	e := Error{
		Source:     "build",
		Message:    "undefined: foo",
		FilePath:   "main.go",
		LineNumber: 42,
		Workspace:  "/projects/myapp",
		Timestamp:  time.Now(),
	}
	if err := db.InsertError(e); err != nil {
		t.Fatalf("InsertError: %v", err)
	}

	rows, err := db.RecentErrors(10)
	if err != nil {
		t.Fatalf("RecentErrors: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 error row, got %d", len(rows))
	}

	got, err := db.GetError(rows[0].ID)
	if err != nil {
		t.Fatalf("GetError: %v", err)
	}
	if got.Source != e.Source {
		t.Errorf("Source: want %q, got %q", e.Source, got.Source)
	}
	if got.FilePath != e.FilePath {
		t.Errorf("FilePath: want %q, got %q", e.FilePath, got.FilePath)
	}
	if got.LineNumber != e.LineNumber {
		t.Errorf("LineNumber: want %d, got %d", e.LineNumber, got.LineNumber)
	}
}

func TestGetError_NotFound(t *testing.T) {
	db := tempDB(t)
	_, err := db.GetError(99999)
	if err == nil {
		t.Error("expected error for missing ID, got nil")
	}
}

func TestRecentErrors_Ordering(t *testing.T) {
	db := tempDB(t)
	for i := 0; i < 5; i++ {
		_ = db.InsertError(Error{
			Source:    "test",
			Message:   "failure",
			Timestamp: time.Now(),
		})
	}
	rows, err := db.RecentErrors(3)
	if err != nil {
		t.Fatalf("RecentErrors: %v", err)
	}
	if len(rows) != 3 {
		t.Errorf("expected 3 rows (limit), got %d", len(rows))
	}
	// Rows should be in descending ID order (most recent first)
	for i := 1; i < len(rows); i++ {
		if rows[i].ID >= rows[i-1].ID {
			t.Errorf("rows not in descending order: %d >= %d", rows[i].ID, rows[i-1].ID)
		}
	}
}

// ---------------------------------------------------------------------------
// Baselines
// ---------------------------------------------------------------------------

func TestBaselineGetSet(t *testing.T) {
	db := tempDB(t)

	// Missing key returns empty string, no error
	v, err := db.GetBaseline("behavior.avg_write_interval_ms")
	if err != nil {
		t.Fatalf("GetBaseline missing key: %v", err)
	}
	if v != "" {
		t.Errorf("expected empty string for missing key, got %q", v)
	}

	if err := db.SetBaseline("behavior.avg_write_interval_ms", "650.00"); err != nil {
		t.Fatalf("SetBaseline: %v", err)
	}

	v, err = db.GetBaseline("behavior.avg_write_interval_ms")
	if err != nil {
		t.Fatalf("GetBaseline: %v", err)
	}
	if v != "650.00" {
		t.Errorf("expected %q, got %q", "650.00", v)
	}

	// Upsert — second write should replace
	if err := db.SetBaseline("behavior.avg_write_interval_ms", "720.00"); err != nil {
		t.Fatalf("SetBaseline upsert: %v", err)
	}
	v, _ = db.GetBaseline("behavior.avg_write_interval_ms")
	if v != "720.00" {
		t.Errorf("upsert: expected %q, got %q", "720.00", v)
	}
}

// ---------------------------------------------------------------------------
// Tripwires
// ---------------------------------------------------------------------------

func TestTripwireLifecycle(t *testing.T) {
	db := tempDB(t)

	tw := Tripwire{
		ID:        "tw-1",
		Pattern:   "payment",
		CreatedAt: time.Now(),
		Active:    true,
	}
	if err := db.InsertTripwire(tw); err != nil {
		t.Fatalf("InsertTripwire: %v", err)
	}

	active, err := db.ActiveTripwires()
	if err != nil {
		t.Fatalf("ActiveTripwires: %v", err)
	}
	if len(active) != 1 || active[0].ID != "tw-1" {
		t.Errorf("expected 1 active tripwire 'tw-1', got %+v", active)
	}

	if err := db.DeactivateTripwire("tw-1"); err != nil {
		t.Fatalf("DeactivateTripwire: %v", err)
	}

	active, _ = db.ActiveTripwires()
	if len(active) != 0 {
		t.Errorf("expected 0 active tripwires after deactivation, got %d", len(active))
	}
}
