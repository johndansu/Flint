package registry

import (
	"testing"
)

func TestLoad_Missing(t *testing.T) {
	dir := t.TempDir()
	r, err := Load(dir)
	if err != nil {
		t.Fatalf("Load on missing file: %v", err)
	}
	if len(r.Workspaces) != 0 {
		t.Errorf("expected empty registry, got %v", r.Workspaces)
	}
}

func TestRegister_AddsWorkspace(t *testing.T) {
	dir := t.TempDir()

	if err := Register(dir, "/projects/alpha"); err != nil {
		t.Fatalf("Register: %v", err)
	}

	r, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(r.Workspaces) != 1 || r.Workspaces[0] != "/projects/alpha" {
		t.Errorf("expected [/projects/alpha], got %v", r.Workspaces)
	}
}

func TestRegister_Idempotent(t *testing.T) {
	dir := t.TempDir()

	for i := 0; i < 3; i++ {
		if err := Register(dir, "/projects/alpha"); err != nil {
			t.Fatalf("Register attempt %d: %v", i, err)
		}
	}

	r, _ := Load(dir)
	if len(r.Workspaces) != 1 {
		t.Errorf("expected 1 workspace after 3 identical registers, got %d", len(r.Workspaces))
	}
}

func TestRegister_MultipleWorkspaces(t *testing.T) {
	dir := t.TempDir()

	workspaces := []string{"/projects/alpha", "/projects/beta", "/projects/gamma"}
	for _, w := range workspaces {
		if err := Register(dir, w); err != nil {
			t.Fatalf("Register %s: %v", w, err)
		}
	}

	r, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(r.Workspaces) != 3 {
		t.Errorf("expected 3 workspaces, got %d: %v", len(r.Workspaces), r.Workspaces)
	}
}

func TestUnregister_RemovesWorkspace(t *testing.T) {
	dir := t.TempDir()

	_ = Register(dir, "/projects/alpha")
	_ = Register(dir, "/projects/beta")

	if err := Unregister(dir, "/projects/alpha"); err != nil {
		t.Fatalf("Unregister: %v", err)
	}

	r, _ := Load(dir)
	if len(r.Workspaces) != 1 || r.Workspaces[0] != "/projects/beta" {
		t.Errorf("expected [/projects/beta] after unregister, got %v", r.Workspaces)
	}
}

func TestUnregister_NoopWhenMissing(t *testing.T) {
	dir := t.TempDir()
	_ = Register(dir, "/projects/alpha")

	// Unregistering something that was never registered should not error
	if err := Unregister(dir, "/projects/nonexistent"); err != nil {
		t.Errorf("Unregister missing: expected no error, got %v", err)
	}

	r, _ := Load(dir)
	if len(r.Workspaces) != 1 {
		t.Errorf("workspace count changed unexpectedly: %v", r.Workspaces)
	}
}

func TestUnregister_EmptyRegistry(t *testing.T) {
	dir := t.TempDir()
	// Should not error even when registry file doesn't exist
	if err := Unregister(dir, "/projects/alpha"); err != nil {
		t.Errorf("Unregister on empty registry: %v", err)
	}
}

func TestRegisterUnregisterAll(t *testing.T) {
	dir := t.TempDir()
	_ = Register(dir, "/a")
	_ = Register(dir, "/b")
	_ = Unregister(dir, "/a")
	_ = Unregister(dir, "/b")

	r, _ := Load(dir)
	if len(r.Workspaces) != 0 {
		t.Errorf("expected empty after unregistering all, got %v", r.Workspaces)
	}
}
