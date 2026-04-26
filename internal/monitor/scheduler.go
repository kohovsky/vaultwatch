package monitor

import (
	"context"
	"log"
	"time"

	"github.com/czeslavo/vaultwatch/internal/alert"
	"github.com/czeslavo/vaultwatch/internal/config"
	"github.com/czeslavo/vaultwatch/internal/vault"
)

// Scheduler periodically polls Vault secrets and triggers alerts.
type Scheduler struct {
	cfg    *config.Config
	client *vault.Client
	hist   *History
	dedup  *DedupWindow
}

// NewScheduler constructs a Scheduler. Returns an error if the history
// directory cannot be initialised.
func NewScheduler(cfg *config.Config, client *vault.Client) (*Scheduler, error) {
	hist, err := NewHistory(cfg.HistoryDir)
	if err != nil {
		return nil, err
	}
	cooldown, _ := time.ParseDuration(cfg.DedupCooldown)
	return &Scheduler{
		cfg:   cfg,
		client: client,
		hist:  hist,
		dedup: NewDedupWindow(cooldown),
	}, nil
}

// Run starts the polling loop and blocks until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context, notifier *alert.Notifier) {
	interval := ParseInterval(s.cfg.Interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("[scheduler] starting — interval=%s", interval)
	s.poll(notifier)

	for {
		select {
		case <-ticker.C:
			s.poll(notifier)
		case <-ctx.Done():
			log.Println("[scheduler] stopping")
			return
		}
	}
}

func (s *Scheduler) poll(notifier *alert.Notifier) {
	infos, err := s.client.GetSecretsInfo(s.cfg.Paths)
	if err != nil {
		log.Printf("[scheduler] vault error: %v", err)
		return
	}

	filtered := Filter(infos, s.cfg.FilterRules)
	statuses := CheckAll(filtered, s.cfg.ParsedThresholds())

	for _, st := range statuses {
		key := DedupKey(st.Path, string(st.Status))
		if s.hist.HasChanged(st.Path, string(st.Status)) || !s.dedup.IsDuplicate(key) {
			if err := notifier.Notify(st); err != nil {
				log.Printf("[scheduler] notify error for %s: %v", st.Path, err)
			}
			s.hist.Record(st.Path, string(st.Status))
		}
	}

	s.dedup.Evict()

	if err := s.hist.Save(); err != nil {
		log.Printf("[scheduler] history save error: %v", err)
	}
}
