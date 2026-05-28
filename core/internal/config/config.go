package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// AwarenessLevel controls how much the daemon watches.
type AwarenessLevel string

const (
	Spark AwarenessLevel = "spark" // manual tools only, no daemon
	Flame AwarenessLevel = "flame" // watches current project
	Forge AwarenessLevel = "forge" // watches all projects
)

// Config is the merged view of global (~/.flint/config.json) and
// workspace-local (.flintrc) settings. Workspace values win.
type Config struct {
	Awareness   AwarenessLevel `json:"awareness"`
	HumanLayer  bool           `json:"humanLayer"`  // opt-in human intelligence
	APIKey      string         `json:"apiKey"`
	Model       string         `json:"model"`
	Project     string         `json:"project"`
	WorkspaceID string         `json:"workspaceId"`
}

const (
	defaultModel = "claude-sonnet-4-20250514"
	globalFile   = "config.json"
	localFile    = ".flintrc"
)

var (
	mu       sync.RWMutex
	current  *Config
	flintDir string // ~/.flint
)

func init() {
	home, _ := os.UserHomeDir()
	flintDir = filepath.Join(home, ".flint")
}

// Load reads global config then overlays workspace-local .flintrc.
// workspaceRoot is the directory containing .flintrc (usually the project root).
func Load(workspaceRoot string) (*Config, error) {
	mu.Lock()
	defer mu.Unlock()

	c := defaults()

	if err := readJSON(filepath.Join(flintDir, globalFile), c); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if workspaceRoot != "" {
		local := &Config{}
		if err := readJSON(filepath.Join(workspaceRoot, localFile), local); err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		overlay(c, local)
	}

	// API key falls back to environment variable
	if c.APIKey == "" {
		c.APIKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	current = c
	return c, nil
}

// Get returns the last loaded config. Panics if Load has not been called.
func Get() *Config {
	mu.RLock()
	defer mu.RUnlock()
	if current == nil {
		panic("config.Load must be called before config.Get")
	}
	return current
}

// SaveGlobal writes c to ~/.flint/config.json (creates the directory if needed).
func SaveGlobal(c *Config) error {
	if err := os.MkdirAll(flintDir, 0o700); err != nil {
		return err
	}
	return writeJSON(filepath.Join(flintDir, globalFile), c)
}

// SaveLocal writes c to workspaceRoot/.flintrc.
func SaveLocal(workspaceRoot string, c *Config) error {
	return writeJSON(filepath.Join(workspaceRoot, localFile), c)
}

// FlintDir returns the path to ~/.flint.
func FlintDir() string { return flintDir }

// defaults returns a Config with safe defaults applied.
func defaults() *Config {
	return &Config{
		Awareness: Spark,
		Model:     defaultModel,
	}
}

// overlay copies non-zero fields from src onto dst.
func overlay(dst, src *Config) {
	if src.Awareness != "" {
		dst.Awareness = src.Awareness
	}
	if src.HumanLayer {
		dst.HumanLayer = src.HumanLayer
	}
	if src.APIKey != "" {
		dst.APIKey = src.APIKey
	}
	if src.Model != "" {
		dst.Model = src.Model
	}
	if src.Project != "" {
		dst.Project = src.Project
	}
	if src.WorkspaceID != "" {
		dst.WorkspaceID = src.WorkspaceID
	}
}

func readJSON(path string, v any) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(v)
}

func writeJSON(path string, v any) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
