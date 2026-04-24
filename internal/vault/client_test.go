package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newMockVaultServer(t *testing.T, leaseDuration int, notFound bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if notFound {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": nil})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"lease_duration": leaseDuration,
			"data": map[string]string{"key": "value"},
		})
	}))
}

func TestNewClient_InvalidAddress(t *testing.T) {
	_, err := NewClient("://bad-address", "token")
	if err == nil {
		t.Fatal("expected error for invalid address, got nil")
	}
}

func TestGetSecretInfo_Success(t *testing.T) {
	server := newMockVaultServer(t, 3600, false)
	defer server.Close()

	client, err := NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}

	info, err := client.GetSecretInfo("secret/data/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Path != "secret/data/myapp" {
		t.Errorf("expected path %q, got %q", "secret/data/myapp", info.Path)
	}

	expectedTTL := 3600 * time.Second
	if info.LeaseTTL != expectedTTL {
		t.Errorf("expected TTL %v, got %v", expectedTTL, info.LeaseTTL)
	}

	if info.ExpiresAt.Before(time.Now()) {
		t.Error("expected ExpiresAt to be in the future")
	}
}

func TestGetSecretsInfo_PartialErrors(t *testing.T) {
	server := newMockVaultServer(t, 7200, false)
	defer server.Close()

	client, err := NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}

	paths := []string{"secret/data/app1", "secret/data/app2"}
	infos, errs := client.GetSecretsInfo(paths)

	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d", len(errs))
	}

	if len(infos) != 2 {
		t.Errorf("expected 2 secret infos, got %d", len(infos))
	}
}
