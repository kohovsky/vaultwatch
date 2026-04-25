package monitor_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/monitor"
	"github.com/yourusername/vaultwatch/internal/vault"
)

func newTestVaultServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"expire_time":"` +
			time.Now().Add(10*time.Minute).UTC().Format(time.RFC3339) +
			`","path":"secret/test"}}`))
	}))
}

func TestScheduler_RunPollsAndStops(t *testing.T) {
	srv := newTestVaultServer(t)
	defer srv.Close()

	var notifyCount int32
	notifyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&notifyCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer notifyServer.Close()

	client, err := vault.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	notifier := alert.NewNotifier(nil)

	cfg := monitor.SchedulerConfig{
		Interval:   50 * time.Millisecond,
		Paths:      []string{"secret/test"},
		Thresholds: []time.Duration{30 * time.Minute, 5 * time.Minute},
		HistoryDir: t.TempDir(),
	}

	sched, err := monitor.NewScheduler(cfg, client, notifier)
	if err != nil {
		t.Fatalf("NewScheduler: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		sched.Run(ctx)
		close(done)
	}()

	select {
	case <-done:
		// scheduler stopped cleanly
	case <-time.After(500 * time.Millisecond):
		t.Fatal("scheduler did not stop after context cancellation")
	}
}

func TestNewScheduler_InvalidHistoryDir(t *testing.T) {
	client, _ := vault.NewClient("http://localhost:8200", "token")
	notifier := alert.NewNotifier(nil)
	cfg := monitor.SchedulerConfig{
		Interval:   time.Minute,
		Paths:      []string{"secret/test"},
		HistoryDir: "/dev/null/invalid/path",
	}
	_, err := monitor.NewScheduler(cfg, client, notifier)
	if err == nil {
		t.Fatal("expected error for invalid history dir, got nil")
	}
}
