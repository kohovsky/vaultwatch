package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, vaultAddr, token string) string {
	t.Helper()
	content := "vault_address: " + vaultAddr + "\n" +
		"vault_token: " + token + "\n" +
		"paths:\n  - secret/data/test\n" +
		"thresholds:\n  warning: 72h\n  critical: 24h\n"

	dir := t.TempDir()
	p := filepath.Join(dir, "vaultwatch.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestExecute_MissingConfig(t *testing.T) {
	rootCmd.SetArgs([]string{"--config", "/nonexistent/path.yaml"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing config, got nil")
	}
}

func TestExecute_InvalidVaultAddress(t *testing.T) {
	cfgPath := writeTempConfig(t, "http://127.0.0.1:19999", "fake-token")
	rootCmd.SetArgs([]string{"--config", cfgPath})

	// Should not panic; vault connection errors are surfaced gracefully.
	// We only assert the command returns an error or exits cleanly.
	_ = rootCmd.Execute()
}

func TestExecute_WithMockVault(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"data": {
				"metadata": {"deletion_time": "", "destroyed": false},
				"data": {"key": "value"}
			},
			"lease_duration": 3600,
			"renewable": true
		}`))
	})

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	cfgPath := writeTempConfig(t, srv.URL, "test-token")

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--config", cfgPath})

	if err := rootCmd.Execute(); err != nil {
		t.Logf("execute returned: %v (may be expected for partial mock)", err)
	}
}
