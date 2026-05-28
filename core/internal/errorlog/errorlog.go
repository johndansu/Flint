package errorlog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"flint/core/internal/store"
)

// Source identifies where the error came from.
type Source string

const (
	SourceStderr    Source = "stderr"
	SourceTest      Source = "test"
	SourceBuild     Source = "build"
	SourceRuntime   Source = "runtime"
	SourcePreCommit Source = "precommit"
	SourceCVE       Source = "cve"
	SourceInstall   Source = "install"
)

// Logger captures errors to SQLite and a human-readable log file.
type Logger struct {
	db        *store.DB
	logPath   string
	workspace string
	mu        sync.Mutex
	file      *os.File
}

// New creates a Logger that writes to db and flintDir/errors.log.
func New(db *store.DB, flintDir, workspace string) (*Logger, error) {
	logPath := filepath.Join(flintDir, "errors.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("errorlog: open %s: %w", logPath, err)
	}
	return &Logger{
		db:        db,
		logPath:   logPath,
		workspace: workspace,
		file:      f,
	}, nil
}

// Close closes the underlying log file.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}

// Log captures a single error entry.
func (l *Logger) Log(source Source, message string, opts ...Option) error {
	e := store.Error{
		Source:    string(source),
		Message:   strings.TrimSpace(message),
		Workspace: l.workspace,
		Timestamp: time.Now(),
	}
	for _, o := range opts {
		o(&e)
	}

	if err := l.db.InsertError(e); err != nil {
		return fmt.Errorf("errorlog: db: %w", err)
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	line := formatLogLine(e)
	_, err := fmt.Fprintln(l.file, line)
	return err
}

// Option configures optional fields on a log entry.
type Option func(*store.Error)

// WithFile attaches a file path to the error.
func WithFile(path string, line int) Option {
	return func(e *store.Error) {
		e.FilePath = path
		e.LineNumber = line
	}
}

// ---------------------------------------------------------------------------
// Convenience constructors
// ---------------------------------------------------------------------------

func (l *Logger) Stderr(message string, opts ...Option) error {
	return l.Log(SourceStderr, message, opts...)
}

func (l *Logger) TestFailure(message string, opts ...Option) error {
	return l.Log(SourceTest, message, opts...)
}

func (l *Logger) BuildError(message string, opts ...Option) error {
	return l.Log(SourceBuild, message, opts...)
}

func (l *Logger) PreCommit(message string, opts ...Option) error {
	return l.Log(SourcePreCommit, message, opts...)
}

func (l *Logger) CVE(message string, opts ...Option) error {
	return l.Log(SourceCVE, message, opts...)
}

// ---------------------------------------------------------------------------
// Formatting
// ---------------------------------------------------------------------------

func formatLogLine(e store.Error) string {
	ts := e.Timestamp.UTC().Format("2006-01-02 15:04:05")
	loc := ""
	if e.FilePath != "" {
		loc = fmt.Sprintf(" [%s", e.FilePath)
		if e.LineNumber > 0 {
			loc += fmt.Sprintf(":%d", e.LineNumber)
		}
		loc += "]"
	}
	return fmt.Sprintf("%s  %-10s  %s%s", ts, e.Source, e.Message, loc)
}

// Tail returns the last n lines of errors.log. Useful for flint status --errors.
func Tail(flintDir string, n int) ([]string, error) {
	logPath := filepath.Join(flintDir, "errors.log")
	b, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(b)), "\n")
	if len(lines) <= n {
		return lines, nil
	}
	return lines[len(lines)-n:], nil
}
