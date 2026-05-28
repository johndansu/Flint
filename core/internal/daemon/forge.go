package daemon

import (
	"context"
	"log"
	"sync"
	"time"

	"flint/core/internal/config"
	"flint/core/internal/ipc"
	"flint/core/internal/registry"
	"flint/core/internal/store"
)

// ForgeDaemon manages one Daemon instance per registered workspace.
// It polls the workspace registry every 30 seconds so newly registered
// workspaces (e.g. from `flint init` in another terminal) are picked up
// without restarting the daemon.
type ForgeDaemon struct {
	cfg       *config.Config
	db        *store.DB
	sharedDir string

	mu     sync.Mutex
	active map[string]context.CancelFunc // root → cancel for that workspace's daemon

	wg sync.WaitGroup
}

// NewForge creates a ForgeDaemon.
func NewForge(cfg *config.Config, db *store.DB, sharedDir string) *ForgeDaemon {
	return &ForgeDaemon{
		cfg:       cfg,
		db:        db,
		sharedDir: sharedDir,
		active:    make(map[string]context.CancelFunc),
	}
}

// Run starts per-workspace daemons for every root in initialRoots, then polls
// the registry every 30 seconds to pick up new workspaces. Blocks until ctx
// is cancelled.
func (f *ForgeDaemon) Run(ctx context.Context, initialRoots []string) error {
	// Load footguns once — shared across all workspace pipelines
	if err := LoadFootguns(f.sharedDir); err != nil {
		log.Printf("forge: footguns: %v (AI-only path)", err)
	}

	for _, root := range initialRoots {
		f.startWorkspace(ctx, root)
	}

	log.Printf("forge: watching %d workspace(s)", len(initialRoots))

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			f.stopAll()
			f.wg.Wait()
			return nil
		case <-ticker.C:
			f.syncRegistry(ctx)
		}
	}
}

// startWorkspace launches a Daemon for root in a background goroutine.
// No-op if that workspace is already active.
func (f *ForgeDaemon) startWorkspace(ctx context.Context, root string) {
	f.mu.Lock()
	if _, ok := f.active[root]; ok {
		f.mu.Unlock()
		return
	}

	socketPath := ipc.SocketPath(config.FlintDir(), root)
	d, err := New(f.cfg, f.db, socketPath, f.sharedDir, root)
	if err != nil {
		log.Printf("forge: skip %s: %v", root, err)
		f.mu.Unlock()
		return
	}

	wsCtx, cancel := context.WithCancel(ctx)
	f.active[root] = cancel
	f.mu.Unlock()

	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		if err := d.Run(wsCtx, root); err != nil && wsCtx.Err() == nil {
			log.Printf("forge: %s exited: %v", root, err)
		}
		f.mu.Lock()
		delete(f.active, root)
		f.mu.Unlock()
	}()

	log.Printf("forge: started watcher for %s", root)
}

// syncRegistry checks the workspace registry and starts watchers for any
// newly registered workspaces that are not yet active.
func (f *ForgeDaemon) syncRegistry(ctx context.Context) {
	reg, err := registry.Load(config.FlintDir())
	if err != nil {
		return
	}
	for _, root := range reg.Workspaces {
		f.startWorkspace(ctx, root)
	}
}

// stopAll cancels every active workspace daemon.
func (f *ForgeDaemon) stopAll() {
	f.mu.Lock()
	defer f.mu.Unlock()
	for root, cancel := range f.active {
		cancel()
		log.Printf("forge: stopping %s", root)
	}
}

// ActiveWorkspaces returns the roots currently being watched.
func (f *ForgeDaemon) ActiveWorkspaces() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, 0, len(f.active))
	for root := range f.active {
		out = append(out, root)
	}
	return out
}
