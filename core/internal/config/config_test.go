package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// writeTestJSON writes v as JSON to path, creating parent dirs as needed.
func writeTestJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(v)
}

// setFlintDir temporarily overrides the global flintDir for testing.
func setFlintDir(t *testing.T, dir string) {
	t.Helper()
	orig := flintDir
	flintDir = dir
	t.Cleanup(func() {
		flintDir = orig
		mu.Lock()
		current = nil
		mu.Unlock()
	})
}

func TestLoad_Defaults(t *testing.T) {
	dir := t.TempDir()
	setFlintDir(t, dir)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Awareness != Spark {
		t.Errorf("default awareness: want spark, got %s", cfg.Awareness)
	}
	if cfg.Model != defaultModel {
		t.Errorf("default model: want %s, got %s", defaultModel, cfg.Model)
	}
}

func TestLoad_GlobalConfig(t *testing.T) {
	dir := t.TempDir()
	setFlintDir(t, dir)

	global := &Config{
		Awareness: Flame,
		APIKey:    "sk-ant-test",
		Model:     "claude-opus-4-7",
	}
	if err := writeTestJSON(filepath.Join(dir, globalFile), global); err != nil {
		t.Fatalf("write global config: %v", err)
	}

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Awareness != Flame {
		t.Errorf("awareness: want flame, got %s", cfg.Awareness)
	}
	if cfg.APIKey != "sk-ant-test" {
		t.Errorf("API key: want sk-ant-test, got %s", cfg.APIKey)
	}
	if cfg.Model != "claude-opus-4-7" {
		t.Errorf("model: want claude-opus-4-7, got %s", cfg.Model)
	}
}

func TestLoad_LocalOverridesGlobal(t *testing.T) {
	dir := t.TempDir()
	wsDir := t.TempDir()
	setFlintDir(t, dir)

	global := &Config{Awareness: Flame, Model: "claude-sonnet-4-6"}
	if err := writeTestJSON(filepath.Join(dir, globalFile), global); err != nil {
		t.Fatal(err)
	}

	// Local overrides awareness only
	local := &Config{Awareness: Forge}
	if err := writeTestJSON(filepath.Join(wsDir, localFile), local); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(wsDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Awareness != Forge {
		t.Errorf("local should override global awareness: want forge, got %s", cfg.Awareness)
	}
	// Non-overridden field comes from global
	if cfg.Model != "claude-sonnet-4-6" {
		t.Errorf("non-overridden model should come from global: want claude-sonnet-4-6, got %s", cfg.Model)
	}
}

func TestLoad_EnvVarOverridesConfig(t *testing.T) {
	dir := t.TempDir()
	setFlintDir(t, dir)

	global := &Config{APIKey: "from-config"}
	if err := writeTestJSON(filepath.Join(dir, globalFile), global); err != nil {
		t.Fatal(err)
	}

	t.Setenv("ANTHROPIC_API_KEY", "from-env")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// Config already has a key, so env var should NOT override it
	if cfg.APIKey != "from-config" {
		t.Errorf("config key should take priority over env: want from-config, got %s", cfg.APIKey)
	}
}

func TestLoad_EnvVarFallback(t *testing.T) {
	dir := t.TempDir()
	setFlintDir(t, dir)

	// No API key in config file
	global := &Config{Awareness: Flame}
	if err := writeTestJSON(filepath.Join(dir, globalFile), global); err != nil {
		t.Fatal(err)
	}

	t.Setenv("ANTHROPIC_API_KEY", "env-key-fallback")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.APIKey != "env-key-fallback" {
		t.Errorf("env var fallback: want env-key-fallback, got %s", cfg.APIKey)
	}
}

func TestLoad_MissingLocalIsOk(t *testing.T) {
	dir := t.TempDir()
	wsDir := t.TempDir()
	setFlintDir(t, dir)

	global := &Config{Awareness: Flame}
	if err := writeTestJSON(filepath.Join(dir, globalFile), global); err != nil {
		t.Fatal(err)
	}

	// wsDir has no .flintrc — should load global cleanly
	cfg, err := Load(wsDir)
	if err != nil {
		t.Fatalf("Load with missing local: %v", err)
	}
	if cfg.Awareness != Flame {
		t.Errorf("want flame from global, got %s", cfg.Awareness)
	}
}

func TestSaveGlobal_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	setFlintDir(t, dir)

	want := &Config{
		Awareness:  Forge,
		HumanLayer: true,
		APIKey:     "sk-ant-xyz",
		Model:      "claude-haiku-4-5",
	}
	if err := SaveGlobal(want); err != nil {
		t.Fatalf("SaveGlobal: %v", err)
	}

	got, err := Load("")
	if err != nil {
		t.Fatalf("Load after save: %v", err)
	}
	if got.Awareness != want.Awareness {
		t.Errorf("Awareness: want %s, got %s", want.Awareness, got.Awareness)
	}
	if got.HumanLayer != want.HumanLayer {
		t.Errorf("HumanLayer: want %v, got %v", want.HumanLayer, got.HumanLayer)
	}
	if got.APIKey != want.APIKey {
		t.Errorf("APIKey: want %s, got %s", want.APIKey, got.APIKey)
	}
}
