package daemon

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileEvent represents a single file system change.
type FileEvent struct {
	Path string
	Op   string // "write" | "create" | "remove" | "rename"
	Time time.Time
}

// Watcher monitors a directory tree for file changes.
type Watcher struct {
	fw     *fsnotify.Watcher
	events chan FileEvent
	errors chan error
}

var watchSkip = map[string]bool{
	"node_modules": true, ".git": true, "vendor": true,
	".cache": true, "dist": true, "build": true, "out": true,
}

var watchExtensions = map[string]bool{
	".go": true, ".ts": true, ".tsx": true, ".js": true, ".jsx": true,
	".py": true, ".rs": true, ".java": true, ".kt": true, ".swift": true,
	".rb": true, ".php": true, ".c": true, ".cpp": true, ".h": true,
	".yaml": true, ".yml": true, ".toml": true, ".json": true,
	".tf": true, ".hcl": true, ".sh": true, ".dockerfile": true,
	".env": true, ".md": true, ".sql": true,
}

// NewWatcher creates and starts a file watcher on root.
func NewWatcher(root string) (*Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		fw:     fw,
		events: make(chan FileEvent, 256),
		errors: make(chan error, 16),
	}

	if err := w.addRecursive(root); err != nil {
		fw.Close()
		return nil, err
	}

	go w.pump()
	return w, nil
}

// Events returns the channel of file events.
func (w *Watcher) Events() <-chan FileEvent { return w.events }

// Errors returns the channel of watcher errors.
func (w *Watcher) Errors() <-chan error { return w.errors }

// Close stops the watcher.
func (w *Watcher) Close() error { return w.fw.Close() }

func (w *Watcher) pump() {
	defer close(w.events)
	defer close(w.errors)

	for {
		select {
		case ev, ok := <-w.fw.Events:
			if !ok {
				return
			}
			// When a new directory is created mid-session, add it to the watcher
			// so files created inside it are also tracked.
			if ev.Op&fsnotify.Create != 0 {
				if info, err := os.Stat(ev.Name); err == nil && info.IsDir() {
					base := filepath.Base(ev.Name)
					if !watchSkip[base] && !strings.HasPrefix(base, ".") {
						_ = w.fw.Add(ev.Name)
					}
					continue // directories themselves are not file events
				}
			}
			if !shouldWatch(ev.Name) {
				continue
			}
			op := opName(ev.Op)
			if op == "" {
				continue
			}
			w.events <- FileEvent{Path: ev.Name, Op: op, Time: time.Now()}

		case err, ok := <-w.fw.Errors:
			if !ok {
				return
			}
			w.errors <- err
		}
	}
}

func (w *Watcher) addRecursive(root string) error {
	// fsnotify doesn't recurse — we add directories manually via filepath.WalkDir.
	// On large repos this is acceptable; we skip heavy directories.
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if watchSkip[base] || strings.HasPrefix(base, ".") && base != "." {
			return filepath.SkipDir
		}
		return w.fw.Add(path)
	})
}

func shouldWatch(path string) bool {
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".") && base != ".env" {
		return false
	}
	ext := strings.ToLower(filepath.Ext(path))
	return watchExtensions[ext]
}

func opName(op fsnotify.Op) string {
	switch {
	case op&fsnotify.Write != 0:
		return "write"
	case op&fsnotify.Create != 0:
		return "create"
	case op&fsnotify.Remove != 0:
		return "remove"
	case op&fsnotify.Rename != 0:
		return "rename"
	}
	return ""
}
