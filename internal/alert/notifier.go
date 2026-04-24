package alert

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/vaultwatch/internal/monitor"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelWarning  Level = "WARNING"
	LevelCritical Level = "CRITICAL"
	LevelExpired  Level = "EXPIRED"
)

// Notifier sends alerts to one or more outputs.
type Notifier struct {
	Writers []io.Writer
}

// NewNotifier returns a Notifier that writes to stdout by default.
func NewNotifier(writers ...io.Writer) *Notifier {
	if len(writers) == 0 {
		writers = []io.Writer{os.Stdout}
	}
	return &Notifier{Writers: writers}
}

// Notify sends an alert message for the given secret result.
func (n *Notifier) Notify(result monitor.SecretStatus) error {
	level := levelFromStatus(result)
	if level == "" {
		return nil
	}
	msg := formatMessage(level, result)
	for _, w := range n.Writers {
		if _, err := fmt.Fprintln(w, msg); err != nil {
			return fmt.Errorf("alert write failed: %w", err)
		}
	}
	return nil
}

// NotifyAll sends alerts for all results that are not OK.
func (n *Notifier) NotifyAll(results []monitor.SecretStatus) []error {
	var errs []error
	for _, r := range results {
		if err := n.Notify(r); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func levelFromStatus(s monitor.SecretStatus) Level {
	switch s.State {
	case monitor.StateWarning:
		return LevelWarning
	case monitor.StateCritical:
		return LevelCritical
	case monitor.StateExpired:
		return LevelExpired
	default:
		return ""
	}
}

func formatMessage(level Level, s monitor.SecretStatus) string {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	return fmt.Sprintf("[%s] %s | path=%s ttl=%s",
		timestamp, level, s.Path, s.TTL.Round(time.Second))
}
