package alert

import (
	"strings"
	"testing"
	"time"

	"github.com/denpeshkov/vaultwatch/internal/monitor"
)

func TestNotify_EmailCalledOnExpired(t *testing.T) {
	var received []string

	// Use a mock writer that captures messages to avoid real SMTP dependency.
	mockWriter := &captureWriter{msgs: &received}

	n := &Notifier{
		writers: []Writer{mockWriter},
	}

	status := monitor.SecretStatus{
		Path:      "secret/db/password",
		Status:    monitor.StatusExpired,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		TTL:       0,
	}

	if err := n.Notify(status); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(received) != 1 {
		t.Fatalf("expected 1 message, got %d", len(received))
	}
	if !strings.Contains(received[0], "secret/db/password") {
		t.Errorf("expected message to contain path, got: %s", received[0])
	}
	if !strings.Contains(received[0], "EXPIRED") {
		t.Errorf("expected message to contain EXPIRED status, got: %s", received[0])
	}
}

// captureWriter is a test helper that records written messages.
type captureWriter struct {
	msgs *[]string
}

func (c *captureWriter) Write(msg string) error {
	*c.msgs = append(*c.msgs, msg)
	return nil
}
