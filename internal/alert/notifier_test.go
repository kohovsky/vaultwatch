package alert_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
	"github.com/vaultwatch/internal/monitor"
)

func makeStatus(state monitor.State, ttl time.Duration) monitor.SecretStatus {
	return monitor.SecretStatus{
		Path:  "secret/data/test",
		TTL:   ttl,
		State: state,
	}
}

func TestNotify_OK_NoOutput(t *testing.T) {
	var buf bytes.Buffer
	n := alert.NewNotifier(&buf)
	s := makeStatus(monitor.StateOK, 48*time.Hour)
	if err := n.Notify(s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for OK state, got: %s", buf.String())
	}
}

func TestNotify_Warning(t *testing.T) {
	var buf bytes.Buffer
	n := alert.NewNotifier(&buf)
	s := makeStatus(monitor.StateWarning, 12*time.Hour)
	if err := n.Notify(s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "WARNING") {
		t.Errorf("expected WARNING in output, got: %s", buf.String())
	}
}

func TestNotify_Critical(t *testing.T) {
	var buf bytes.Buffer
	n := alert.NewNotifier(&buf)
	s := makeStatus(monitor.StateCritical, 1*time.Hour)
	if err := n.Notify(s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "CRITICAL") {
		t.Errorf("expected CRITICAL in output, got: %s", buf.String())
	}
}

func TestNotify_Expired(t *testing.T) {
	var buf bytes.Buffer
	n := alert.NewNotifier(&buf)
	s := makeStatus(monitor.StateExpired, 0)
	if err := n.Notify(s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "EXPIRED") {
		t.Errorf("expected EXPIRED in output, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "secret/data/test") {
		t.Errorf("expected path in output, got: %s", buf.String())
	}
}

func TestNotifyAll_MultipleWriters(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	n := alert.NewNotifier(&buf1, &buf2)
	results := []monitor.SecretStatus{
		makeStatus(monitor.StateOK, 72*time.Hour),
		makeStatus(monitor.StateWarning, 6*time.Hour),
		makeStatus(monitor.StateCritical, 30*time.Minute),
	}
	errs := n.NotifyAll(results)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	for _, buf := range []*bytes.Buffer{&buf1, &buf2} {
		out := buf.String()
		if !strings.Contains(out, "WARNING") {
			t.Errorf("expected WARNING in output")
		}
		if !strings.Contains(out, "CRITICAL") {
			t.Errorf("expected CRITICAL in output")
		}
	}
}
