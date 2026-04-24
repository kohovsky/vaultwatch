package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/your-org/vaultwatch/internal/config"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "vaultwatch-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

const validConfig = `
vault:
  address: "https://vault.example.com"
  token: "s.testtoken"
monitor:
  interval: "5m"
  paths:
    - "secret/data/myapp"
alerts:
  thresholds: ["72h", "24h", "1h"]
  slack:
    webhook_url: "https://hooks.slack.com/test"
    channel: "#alerts"
`

func TestLoad_ValidConfig(t *testing.T) {
	path := writeTempConfig(t, validConfig)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Vault.Address != "https://vault.example.com" {
		t.Errorf("unexpected vault address: %s", cfg.Vault.Address)
	}
	if len(cfg.Monitor.Paths) != 1 {
		t.Errorf("expected 1 path, got %d", len(cfg.Monitor.Paths))
	}
}

func TestLoad_MissingVaultAddress(t *testing.T) {
	content := `
vault:
  token: "s.testtoken"
monitor:
  paths: ["secret/data/app"]
`
	path := writeTempConfig(t, content)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error for missing vault address")
	}
}

func TestLoad_MissingPaths(t *testing.T) {
	content := `
vault:
  address: "https://vault.example.com"
  token: "s.testtoken"
monitor:
  paths: []
`
	path := writeTempConfig(t, content)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error for empty paths")
	}
}

func TestParsedThresholds(t *testing.T) {
	path := writeTempConfig(t, validConfig)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	thresholds, err := cfg.Alerts.ParsedThresholds()
	if err != nil {
		t.Fatalf("unexpected error parsing thresholds: %v", err)
	}
	expected := []time.Duration{72 * time.Hour, 24 * time.Hour, 1 * time.Hour}
	for i, d := range thresholds {
		if d != expected[i] {
			t.Errorf("threshold[%d]: expected %v, got %v", i, expected[i], d)
		}
	}
}

func TestParsedThresholds_InvalidDuration(t *testing.T) {
	cfg := &config.Config{}
	cfg.Alerts.Thresholds = []string{"not-a-duration"}
	_, err := cfg.Alerts.ParsedThresholds()
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
}
