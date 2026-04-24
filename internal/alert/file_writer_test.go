package alert_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vaultwatch/internal/alert"
)

func TestFileWriter_WritesContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "alerts", "vaultwatch.log")

	fw, err := alert.NewFileWriter(path)
	if err != nil {
		t.Fatalf("NewFileWriter: %v", err)
	}
	defer fw.Close()

	msg := "[2024-01-01T00:00:00Z] WARNING | path=secret/data/db ttl=11h0m0s\n"
	if _, err := fw.Write([]byte(msg)); err != nil {
		t.Fatalf("Write: %v", err)
	}
	fw.Close()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(data), "WARNING") {
		t.Errorf("expected WARNING in file, got: %s", string(data))
	}
}

func TestFileWriter_CreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "deep", "alerts.log")

	fw, err := alert.NewFileWriter(path)
	if err != nil {
		t.Fatalf("unexpected error creating nested dirs: %v", err)
	}
	fw.Close()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file to exist at %s", path)
	}
}

func TestFileWriter_AppendsBetweenOpens(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "append.log")

	for _, line := range []string{"line1\n", "line2\n"} {
		fw, err := alert.NewFileWriter(path)
		if err != nil {
			t.Fatalf("NewFileWriter: %v", err)
		}
		if _, err := fw.Write([]byte(line)); err != nil {
			t.Fatalf("Write: %v", err)
		}
		fw.Close()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(data), "line1") || !strings.Contains(string(data), "line2") {
		t.Errorf("expected both lines in file, got: %s", string(data))
	}
}
