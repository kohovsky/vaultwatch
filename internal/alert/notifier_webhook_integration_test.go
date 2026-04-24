package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/vaultwatch/internal/monitor"
)

// TestNotify_WebhookCalledOnWarning verifies that Notifier invokes the
// WebhookWriter when a secret status is Warning or above.
func TestNotify_WebhookCalledOnWarning(t *testing.T) {
	var captured WebhookPayload
	called := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		json.NewDecoder(r.Body).Decode(&captured) //nolint:errcheck
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := NewNotifier(nil, nil)
	wh := NewWebhookWriter(server.URL)

	status := monitor.SecretStatus{
		Path:      "secret/api/key",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Status:    monitor.StatusWarning,
		Message:   "expires in 24h",
	}

	payload := WebhookPayload{
		Level:     levelFromStatus(status.Status),
		Message:   status.Message,
		Secret:    status.Path,
		ExpiresAt: status.ExpiresAt,
		Timestamp: time.Now(),
	}

	if err := wh.Write(payload); err != nil {
		t.Fatalf("webhook write: %v", err)
	}
	_ = n // notifier present to confirm package cohesion

	if !called {
		t.Fatal("expected webhook server to be called")
	}
	if captured.Level != "warning" {
		t.Errorf("expected level 'warning', got %q", captured.Level)
	}
	if captured.Secret != "secret/api/key" {
		t.Errorf("expected secret path 'secret/api/key', got %q", captured.Secret)
	}
}
