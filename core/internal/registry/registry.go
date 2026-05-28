package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Registry struct {
	Workspaces []string `json:"workspaces"`
}

// Path returns the location of the registry file.
func Path(flintDir string) string {
	return filepath.Join(flintDir, "workspaces.json")
}

// Load reads the registry. Returns an empty registry if the file does not exist.
func Load(flintDir string) (*Registry, error) {
	b, err := os.ReadFile(Path(flintDir))
	if os.IsNotExist(err) {
		return &Registry{}, nil
	}
	if err != nil {
		return nil, err
	}
	var r Registry
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// Register adds workspaceRoot to the registry if it is not already present.
func Register(flintDir, workspaceRoot string) error {
	r, _ := Load(flintDir)
	if r == nil {
		r = &Registry{}
	}
	for _, w := range r.Workspaces {
		if w == workspaceRoot {
			return nil
		}
	}
	r.Workspaces = append(r.Workspaces, workspaceRoot)
	return save(flintDir, r)
}

// Unregister removes workspaceRoot from the registry.
func Unregister(flintDir, workspaceRoot string) error {
	r, _ := Load(flintDir)
	if r == nil {
		return nil
	}
	filtered := r.Workspaces[:0]
	for _, w := range r.Workspaces {
		if w != workspaceRoot {
			filtered = append(filtered, w)
		}
	}
	r.Workspaces = filtered
	return save(flintDir, r)
}

func save(flintDir string, r *Registry) error {
	if err := os.MkdirAll(flintDir, 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(Path(flintDir), b, 0o600)
}
