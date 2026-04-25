package monitor

import (
	"context"
	"log"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/vault"
)

// SchedulerConfig holds configuration for the polling scheduler.
type SchedulerConfig struct {
	Interval   time.Duration
	Paths      []string
	Thresholds []time.Duration
	HistoryDir string
}

// Scheduler periodically polls Vault secrets and dispatches alerts.
type Scheduler struct {
	cfg      SchedulerConfig
	client   *vault.Client
	notifier *alert.Notifier
	history  *History
}

// NewScheduler creates a new Scheduler.
func NewScheduler(cfg SchedulerConfig, client *vault.Client, notifier *alert.Notifier) (*Scheduler, error) {
	h, err := NewHistory(cfg.HistoryDir)
	if err != nil {
		return nil, err
	}
	return &Scheduler{
		cfg:      cfg,
		client:   client,
		notifier: notifier,
		history:  h,
	}, nil
}

// Run starts the scheduler loop, blocking until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.Interval)
	defer ticker.Stop()

	log.Printf("[scheduler] starting poll loop every %s", s.cfg.Interval)
	s.poll(ctx)

	for {
		select {
		case <-ticker.C:
			s.poll(ctx)
		case <-ctx.Done():
			log.Println("[scheduler] context cancelled, stopping")
			return
		}
	}
}

func (s *Scheduler) poll(ctx context.Context) {
	statuses := CheckAll(ctx, s.client, s.cfg.Paths, s.cfg.Thresholds)
	for _, st := range statuses {
		if s.history.HasChanged(st.Path, st.Status) {
			if err := s.notifier.Notify(st); err != nil {
				log.Printf("[scheduler] notify error for %s: %v", st.Path, err)
			}
			s.history.Record(st.Path, st.Status)
		}
	}
	if err := s.history.Save(); err != nil {
		log.Printf("[scheduler] failed to save history: %v", err)
	}
	summary := Summarize(statuses)
	log.Printf("[scheduler] poll complete: %s", FormatSummary(summary))
}
