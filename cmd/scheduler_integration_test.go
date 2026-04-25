package cmd_test

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

func TestSchedulerIntegration_AlertOnStatusChange(t *testing.T) {
	var callCount int32

	vaultSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"expire_time":"` +
			time.Now().Add(2*time.Minute).UTC().Format(time.RFC3339) +
			`","path":"secret/db"}}`))
	}))
	defer vaultSrv.Close()

	alertSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer alertSrv.Close()

	client, err := vault.NewClient(vaultSrv.URL, "root")
	if err != nil {
		t.Fatalf("vault client: %v", err)
	}

	webhook, err := alert.NewWebhookWriter(alertSrv.URL)
	if err != nil {
		t.Fatalf("webhook writer: %v", err)
	}
	notifier := alert.NewNotifier([]alert.Writer{webhook})

	cfg := monitor.SchedulerConfig{
		Interval:   40 * time.Millisecond,
		Paths:      []string{"secret/db"},
		Thresholds: []time.Duration{10 * time.Minute, 1 * time.Minute},
		HistoryDir: t.TempDir(),
	}

	sched, err := monitor.NewScheduler(cfg, client, notifier)
	if err != nil {
		t.Fatalf("NewScheduler: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	sched.Run(ctx)

	// First poll triggers alert (new path); subsequent polls skip (no status change).
	if atomic.LoadInt32(&callCount) < 1 {
		t.Errorf("expected at least 1 webhook call, got %d", callCount)
	}
}
