package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// DB wraps the SQLite connection for Flint's local data store.
type DB struct {
	sql *sql.DB
}

// Open opens (or creates) the Flint SQLite database at flintDir/flint.db
// and runs all schema migrations.
func Open(flintDir string) (*DB, error) {
	if err := os.MkdirAll(flintDir, 0o700); err != nil {
		return nil, fmt.Errorf("store: mkdir %s: %w", flintDir, err)
	}

	path := filepath.Join(flintDir, "flint.db")
	db, err := sql.Open("sqlite", path+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("store: open %s: %w", path, err)
	}
	db.SetMaxOpenConns(1) // SQLite: single writer

	s := &DB{sql: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// Close closes the underlying database connection.
func (d *DB) Close() error { return d.sql.Close() }

// Raw returns the underlying *sql.DB for callers that need direct access.
func (d *DB) Raw() *sql.DB { return d.sql }

// ---------------------------------------------------------------------------
// Schema
// ---------------------------------------------------------------------------

func (d *DB) migrate() error {
	if _, err := d.sql.Exec(schema); err != nil {
		return err
	}
	return d.runMigrations()
}

// runMigrations applies incremental ALTER TABLE changes.
// Each statement is run independently so a duplicate-column error on an
// already-migrated database is silently ignored.
func (d *DB) runMigrations() error {
	alters := []string{
		`ALTER TABLE sessions ADD COLUMN last_activity_at TEXT`,
	}
	for _, stmt := range alters {
		if _, err := d.sql.Exec(stmt); err != nil {
			if !strings.Contains(err.Error(), "duplicate column name") {
				return fmt.Errorf("store: migration %q: %w", stmt, err)
			}
		}
	}
	return nil
}

const schema = `
CREATE TABLE IF NOT EXISTS observations (
	id            TEXT PRIMARY KEY,
	kind          TEXT NOT NULL,
	category      TEXT NOT NULL,
	confidence    REAL NOT NULL,
	text          TEXT NOT NULL,
	agent_prompt  TEXT,
	explain_why   TEXT,
	file_path     TEXT,
	line_number   INTEGER,
	session_type  TEXT,
	footgun_id    TEXT,
	timestamp     TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS dismissals (
	id              INTEGER PRIMARY KEY AUTOINCREMENT,
	observation_id  TEXT NOT NULL,
	action          TEXT NOT NULL,  -- 'calibrated' | 'dismissed'
	timestamp       TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	session_type  TEXT NOT NULL,
	state         TEXT NOT NULL,
	started_at    TEXT NOT NULL,
	ended_at      TEXT,
	workspace     TEXT
);

CREATE TABLE IF NOT EXISTS errors (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	source      TEXT NOT NULL,  -- 'stderr' | 'test' | 'build' | 'runtime' | 'precommit' | 'cve' | 'install'
	message     TEXT NOT NULL,
	file_path   TEXT,
	line_number INTEGER,
	workspace   TEXT,
	timestamp   TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS auto_fixes (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	category      TEXT NOT NULL,
	description   TEXT NOT NULL,
	files_changed TEXT NOT NULL,  -- JSON array
	staged        INTEGER NOT NULL DEFAULT 1,
	timestamp     TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS footgun_hits (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	footgun_id  TEXT NOT NULL,
	file_path   TEXT NOT NULL,
	line_number INTEGER,
	context     TEXT,
	timestamp   TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS tripwires (
	id         TEXT PRIMARY KEY,
	pattern    TEXT NOT NULL UNIQUE,
	created_at TEXT NOT NULL,
	active     INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS baselines (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	key        TEXT NOT NULL UNIQUE,
	value      TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS tool_usage (
	tool_id    TEXT NOT NULL,
	role       TEXT NOT NULL,
	used_at    TEXT NOT NULL,
	PRIMARY KEY (tool_id, used_at)
);

CREATE INDEX IF NOT EXISTS idx_observations_timestamp  ON observations(timestamp);
CREATE INDEX IF NOT EXISTS idx_observations_kind       ON observations(kind);
CREATE INDEX IF NOT EXISTS idx_errors_timestamp        ON errors(timestamp);
CREATE INDEX IF NOT EXISTS idx_errors_source           ON errors(source);
CREATE INDEX IF NOT EXISTS idx_footgun_hits_footgun_id ON footgun_hits(footgun_id);
`

// ---------------------------------------------------------------------------
// Observations
// ---------------------------------------------------------------------------

type Observation struct {
	ID           string
	Kind         string
	Category     string
	Confidence   float64
	Text         string
	AgentPrompt  string
	ExplainWhy   string
	FilePath     string
	LineNumber   int
	SessionType  string
	FootgunID    string
	Timestamp    time.Time
}

func (d *DB) InsertObservation(o Observation) error {
	_, err := d.sql.Exec(`
		INSERT OR IGNORE INTO observations
		(id, kind, category, confidence, text, agent_prompt, explain_why,
		 file_path, line_number, session_type, footgun_id, timestamp)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
		o.ID, o.Kind, o.Category, o.Confidence, o.Text,
		nullStr(o.AgentPrompt), nullStr(o.ExplainWhy),
		nullStr(o.FilePath), nullInt(o.LineNumber),
		nullStr(o.SessionType), nullStr(o.FootgunID),
		o.Timestamp.UTC().Format(time.RFC3339),
	)
	return err
}

// ---------------------------------------------------------------------------
// Sessions
// ---------------------------------------------------------------------------

// StartSession inserts a new session row and returns its ID.
func (d *DB) StartSession(sessionType, workspace string) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := d.sql.Exec(`
		INSERT INTO sessions (session_type, state, started_at, last_activity_at, workspace)
		VALUES (?, 'active', ?, ?, ?)`,
		sessionType, now, now, nullStr(workspace),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// EndSession marks a session as ended.
func (d *DB) EndSession(id int64) error {
	_, err := d.sql.Exec(`
		UPDATE sessions SET ended_at = ?, state = 'ended' WHERE id = ?`,
		time.Now().UTC().Format(time.RFC3339), id,
	)
	return err
}

// ---------------------------------------------------------------------------
// Errors
// ---------------------------------------------------------------------------

type Error struct {
	Source     string
	Message    string
	FilePath   string
	LineNumber int
	Workspace  string
	Timestamp  time.Time
}

// ErrorRow is an Error with its database ID included.
type ErrorRow struct {
	ID int64
	Error
}

func (d *DB) InsertError(e Error) error {
	_, err := d.sql.Exec(`
		INSERT INTO errors (source, message, file_path, line_number, workspace, timestamp)
		VALUES (?,?,?,?,?,?)`,
		e.Source, e.Message,
		nullStr(e.FilePath), nullInt(e.LineNumber),
		nullStr(e.Workspace),
		e.Timestamp.UTC().Format(time.RFC3339),
	)
	return err
}

func (d *DB) GetError(id int64) (*Error, error) {
	var e Error
	var fp, ws, ts sql.NullString
	var ln sql.NullInt64
	err := d.sql.QueryRow(`
		SELECT source, message, file_path, line_number, workspace, timestamp
		FROM errors WHERE id = ?`, id,
	).Scan(&e.Source, &e.Message, &fp, &ln, &ws, &ts)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("error #%d not found", id)
	}
	if err != nil {
		return nil, err
	}
	if fp.Valid {
		e.FilePath = fp.String
	}
	if ln.Valid {
		e.LineNumber = int(ln.Int64)
	}
	if ws.Valid {
		e.Workspace = ws.String
	}
	if ts.Valid {
		e.Timestamp, _ = time.Parse(time.RFC3339, ts.String)
	}
	return &e, nil
}

func (d *DB) RecentErrors(limit int) ([]ErrorRow, error) {
	rows, err := d.sql.Query(`
		SELECT id, source, message, file_path, line_number, timestamp
		FROM errors ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ErrorRow
	for rows.Next() {
		var r ErrorRow
		var fp, ts sql.NullString
		var ln sql.NullInt64
		if err := rows.Scan(&r.ID, &r.Source, &r.Message, &fp, &ln, &ts); err != nil {
			return nil, err
		}
		if fp.Valid {
			r.FilePath = fp.String
		}
		if ln.Valid {
			r.LineNumber = int(ln.Int64)
		}
		if ts.Valid {
			r.Timestamp, _ = time.Parse(time.RFC3339, ts.String)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ---------------------------------------------------------------------------
// Baselines
// ---------------------------------------------------------------------------

func (d *DB) GetBaseline(key string) (string, error) {
	var value string
	err := d.sql.QueryRow(`SELECT value FROM baselines WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (d *DB) SetBaseline(key, value string) error {
	_, err := d.sql.Exec(`
		INSERT INTO baselines (key, value, updated_at) VALUES (?,?,?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at`,
		key, value, time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

// ---------------------------------------------------------------------------
// Observations (queries)
// ---------------------------------------------------------------------------

func (d *DB) CountObservations() (int, error) {
	var n int
	err := d.sql.QueryRow(`SELECT COUNT(*) FROM observations`).Scan(&n)
	return n, err
}

// GetObservationByPrefix finds the observation whose ID starts with prefix.
// Returns an error if zero or more than one match is found.
func (d *DB) GetObservationByPrefix(prefix string) (*Observation, error) {
	rows, err := d.sql.Query(`
		SELECT id, kind, category, confidence, text, file_path, line_number, session_type, timestamp
		FROM observations WHERE id LIKE ? LIMIT 2`, prefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []Observation
	for rows.Next() {
		var o Observation
		var fp, st, ts sql.NullString
		var ln sql.NullInt64
		if err := rows.Scan(&o.ID, &o.Kind, &o.Category, &o.Confidence, &o.Text, &fp, &ln, &st, &ts); err != nil {
			return nil, err
		}
		if fp.Valid {
			o.FilePath = fp.String
		}
		if ln.Valid {
			o.LineNumber = int(ln.Int64)
		}
		if st.Valid {
			o.SessionType = st.String
		}
		if ts.Valid {
			o.Timestamp, _ = time.Parse(time.RFC3339, ts.String)
		}
		matches = append(matches, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("no observation with ID prefix %q", prefix)
	case 1:
		return &matches[0], nil
	default:
		return nil, fmt.Errorf("ambiguous prefix %q — be more specific", prefix)
	}
}

// InsertDismissal records a dismiss or calibrate action for an observation.
// action should be "dismissed" or "calibrated".
func (d *DB) InsertDismissal(observationID, action string) error {
	_, err := d.sql.Exec(`
		INSERT INTO dismissals (observation_id, action, timestamp)
		VALUES (?,?,?)`,
		observationID, action, time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

// ObservationKindCounts returns a map of kind → count.
func (d *DB) ObservationKindCounts() (map[string]int, error) {
	rows, err := d.sql.Query(`SELECT kind, COUNT(*) FROM observations GROUP BY kind`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]int)
	for rows.Next() {
		var kind string
		var n int
		if err := rows.Scan(&kind, &n); err != nil {
			return nil, err
		}
		out[kind] = n
	}
	return out, rows.Err()
}

// RecentObservations returns the N most recent observations.
func (d *DB) RecentObservations(limit int) ([]Observation, error) {
	rows, err := d.sql.Query(`
		SELECT id, kind, category, confidence, text, file_path, line_number, timestamp
		FROM observations ORDER BY timestamp DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Observation
	for rows.Next() {
		var o Observation
		var fp, ts sql.NullString
		var ln sql.NullInt64
		var conf float64
		if err := rows.Scan(&o.ID, &o.Kind, &o.Category, &conf, &o.Text, &fp, &ln, &ts); err != nil {
			return nil, err
		}
		o.Confidence = conf
		if fp.Valid {
			o.FilePath = fp.String
		}
		if ln.Valid {
			o.LineNumber = int(ln.Int64)
		}
		if ts.Valid {
			o.Timestamp, _ = time.Parse(time.RFC3339, ts.String)
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// AllBaselines returns all baselines sorted by key.
func (d *DB) AllBaselines() ([][2]string, error) {
	rows, err := d.sql.Query(`SELECT key, value FROM baselines ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out [][2]string
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		out = append(out, [2]string{k, v})
	}
	return out, rows.Err()
}

// ClearObservations deletes all rows from the observations table.
func (d *DB) ClearObservations() error {
	_, err := d.sql.Exec(`DELETE FROM observations`)
	return err
}

// FilteredErrors returns errors optionally filtered by source, newest first.
// limit=0 means no limit.
func (d *DB) FilteredErrors(filterType string, limit int) ([]ErrorRow, error) {
	query := `SELECT id, source, message, file_path, line_number, timestamp FROM errors`
	args := []any{}
	if filterType != "" {
		query += ` WHERE source = ?`
		args = append(args, filterType)
	}
	query += ` ORDER BY id DESC`
	if limit > 0 {
		query += fmt.Sprintf(` LIMIT %d`, limit)
	}

	rows, err := d.sql.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ErrorRow
	for rows.Next() {
		var r ErrorRow
		var fp, ts sql.NullString
		var ln sql.NullInt64
		if err := rows.Scan(&r.ID, &r.Source, &r.Message, &fp, &ln, &ts); err != nil {
			return nil, err
		}
		if fp.Valid {
			r.FilePath = fp.String
		}
		if ln.Valid {
			r.LineNumber = int(ln.Int64)
		}
		if ts.Valid {
			r.Timestamp, _ = time.Parse(time.RFC3339, ts.String)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// HumanObservations returns the most recent observations with kind "human".
func (d *DB) HumanObservations(limit int) ([]Observation, error) {
	rows, err := d.sql.Query(`
		SELECT id, kind, category, confidence, text, file_path, line_number, timestamp
		FROM observations WHERE kind = 'human'
		ORDER BY timestamp DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Observation
	for rows.Next() {
		var o Observation
		var fp, ts sql.NullString
		var ln sql.NullInt64
		if err := rows.Scan(&o.ID, &o.Kind, &o.Category, &o.Confidence, &o.Text, &fp, &ln, &ts); err != nil {
			return nil, err
		}
		if fp.Valid {
			o.FilePath = fp.String
		}
		if ln.Valid {
			o.LineNumber = int(ln.Int64)
		}
		if ts.Valid {
			o.Timestamp, _ = time.Parse(time.RFC3339, ts.String)
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// ClearHumanObservations deletes all observations with kind "human".
func (d *DB) ClearHumanObservations() error {
	_, err := d.sql.Exec(`DELETE FROM observations WHERE kind = 'human'`)
	return err
}

// ClearErrors deletes all rows from the errors table.
func (d *DB) ClearErrors() error {
	_, err := d.sql.Exec(`DELETE FROM errors`)
	return err
}

// ClearAll deletes all transient data: observations, errors, dismissals, auto_fixes.
func (d *DB) ClearAll() error {
	for _, tbl := range []string{"observations", "errors", "dismissals", "auto_fixes"} {
		if _, err := d.sql.Exec(`DELETE FROM ` + tbl); err != nil {
			return fmt.Errorf("clear %s: %w", tbl, err)
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Tripwires
// ---------------------------------------------------------------------------

type Tripwire struct {
	ID        string
	Pattern   string
	CreatedAt time.Time
	Active    bool
}

func (d *DB) InsertTripwire(t Tripwire) error {
	_, err := d.sql.Exec(`
		INSERT INTO tripwires (id, pattern, created_at, active)
		VALUES (?,?,?,?)`,
		t.ID, t.Pattern, t.CreatedAt.UTC().Format(time.RFC3339), boolInt(t.Active),
	)
	return err
}

func (d *DB) ActiveTripwires() ([]Tripwire, error) {
	rows, err := d.sql.Query(`SELECT id, pattern, created_at FROM tripwires WHERE active = 1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Tripwire
	for rows.Next() {
		var t Tripwire
		var ts string
		if err := rows.Scan(&t.ID, &t.Pattern, &ts); err != nil {
			return nil, err
		}
		t.CreatedAt, _ = time.Parse(time.RFC3339, ts)
		t.Active = true
		out = append(out, t)
	}
	return out, rows.Err()
}

func (d *DB) DeactivateTripwire(id string) error {
	_, err := d.sql.Exec(`UPDATE tripwires SET active = 0 WHERE id = ?`, id)
	return err
}

// ---------------------------------------------------------------------------
// Auto-fixes
// ---------------------------------------------------------------------------

type AutoFix struct {
	Category     string
	Description  string
	FilesChanged string // JSON-encoded array
	Timestamp    time.Time
}

func (d *DB) InsertAutoFix(f AutoFix) error {
	_, err := d.sql.Exec(`
		INSERT INTO auto_fixes (category, description, files_changed, timestamp)
		VALUES (?,?,?,?)`,
		f.Category, f.Description, f.FilesChanged,
		f.Timestamp.UTC().Format(time.RFC3339),
	)
	return err
}

// ---------------------------------------------------------------------------
// Tool usage
// ---------------------------------------------------------------------------

func (d *DB) RecordToolUsage(toolID, role string) error {
	_, err := d.sql.Exec(`
		INSERT OR IGNORE INTO tool_usage (tool_id, role, used_at) VALUES (?,?,?)`,
		toolID, role, time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullInt(n int) any {
	if n == 0 {
		return nil
	}
	return n
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
