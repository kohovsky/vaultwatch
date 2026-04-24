package alert

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

func TestNotify_SlackCalledOnCritical(t *testing.T) {
	var capturedText string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload slackPayload
		_ = json.Unmarshal(body, &payload)
		capturedText = payload.Text
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	slack := NewSlackWriter(server.URL)
	notifier := NewNotifier([]Writer{slack})

	status := monitor.SecretStatus{
		Path:      "secret/db/prod",
		ExpiresAt: time.Now().Add(2 * time.Hour),
		Status:    monitor.StatusCritical,
		Message:   "expires in 2h0m0s",
	}

	if err := notifier.Notify(status); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(capturedText, "secret/db/prod") {
		t.Errorf("expected path in slack message, got: %q", capturedText)
	}
	if !strings.Contains(capturedText, "CRITICAL") {
		t.Errorf("expected CRITICAL level in slack message, got: %q", capturedText)
	}
}
