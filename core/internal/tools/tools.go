package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Tool is one entry from shared/tools.json.
type Tool struct {
	ID          string `json:"id"`
	Role        string `json:"role"`
	Category    string `json:"category"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Placeholder string `json:"placeholder"`
	Hint        string `json:"hint"`
	Prompt      string `json:"prompt"`
}

type toolsFile struct {
	Tools []Tool `json:"tools"`
}

var registry []Tool

// Load reads shared/tools.json relative to the binary's location.
// sharedDir is the path to the shared/ directory.
func Load(sharedDir string) error {
	p := filepath.Join(sharedDir, "tools.json")
	f, err := os.Open(p)
	if err != nil {
		return fmt.Errorf("tools: open %s: %w", p, err)
	}
	defer f.Close()

	var tf toolsFile
	if err := json.NewDecoder(f).Decode(&tf); err != nil {
		return fmt.Errorf("tools: parse %s: %w", p, err)
	}
	registry = tf.Tools
	return nil
}

// All returns every loaded tool.
func All() []Tool { return registry }

// ByID looks up a tool by its id field. Returns an error if not found.
func ByID(id string) (Tool, error) {
	for _, t := range registry {
		if t.ID == id {
			return t, nil
		}
	}
	return Tool{}, fmt.Errorf("tools: unknown tool %q", id)
}

// ForRole returns all tools for a given role.
func ForRole(role string) []Tool {
	var out []Tool
	for _, t := range registry {
		if t.Role == role {
			out = append(out, t)
		}
	}
	return out
}
