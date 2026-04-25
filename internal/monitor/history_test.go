package monitor

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHistory_HasChanged_NewPath(t *testing.T) {
	h, err := NewHistory("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !h.HasChanged("secret/foo", "warning") {
		t.Error("expected HasChanged=true for unseen path")
	}
}

func TestHistory_HasChanged_SameStatus(t *testing.T) {
	h, _ := NewHistory("")
	h.Record("secret/foo", "warning", time.Now().Add(time.Hour))
	if h.HasChanged("secret/foo", "warning") {
		t.Error("expected HasChanged=false when status unchanged")
	}
}

func TestHistory_HasChanged_DifferentStatus(t *testing.T) {
	h, _ := NewHistory("")
	h.Record("secret/foo", "ok", time.Now().Add(time.Hour))
	if !h.HasChanged("secret/foo", "critical") {
		t.Error("expected HasChanged=true when status changed")
	}
}

func TestHistory_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	h1, err := NewHistory(path)
	if err != nil {
		t.Fatalf("NewHistory: %v", err)
	}
	expiry := time.Now().Add(2 * time.Hour).Truncate(time.Second)
	h1.Record("secret/bar", "critical", expiry)
	if err := h1.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	h2, err := NewHistory(path)
	if err != nil {
		t.Fatalf("reload NewHistory: %v", err)
	}
	if h2.HasChanged("secret/bar", "critical") {
		t.Error("expected HasChanged=false after reload with same status")
	}
	if !h2.HasChanged("secret/bar", "ok") {
		t.Error("expected HasChanged=true after reload with different status")
	}
}

func TestHistory_NoFile_NoOp(t *testing.T) {
	h, _ := NewHistory("")
	if err := h.Save(); err != nil {
		t.Errorf("Save with empty path should be no-op, got: %v", err)
	}
}

func TestHistory_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")
	h, _ := NewHistory("")
	h.filePath = path
	h.Record("secret/x", "ok", time.Now().Add(time.Hour))
	if err := h.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected file mode 0600, got %v", info.Mode().Perm())
	}
}
