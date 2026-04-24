package monitor

import (
	"fmt"
	"time"

	"github.com/user/vaultwatch/internal/vault"
)

// ExpiryStatus represents the expiration state of a secret.
type ExpiryStatus int

const (
	StatusOK      ExpiryStatus = iota // Secret is healthy
	StatusWarning                     // Secret is within a warning threshold
	StatusCritical                    // Secret is within a critical threshold
	StatusExpired                     // Secret has already expired
)

// SecretAlert holds alert information for a secret approaching expiry.
type SecretAlert struct {
	Path      string
	Status    ExpiryStatus
	ExpiresAt time.Time
	TimeLeft  time.Duration
}

// String returns a human-readable representation of the alert status.
func (s ExpiryStatus) String() string {
	switch s {
	case StatusOK:
		return "OK"
	case StatusWarning:
		return "WARNING"
	case StatusCritical:
		return "CRITICAL"
	case StatusExpired:
		return "EXPIRED"
	default:
		return "UNKNOWN"
	}
}

// CheckExpiry evaluates a secret's expiration against the provided thresholds.
// warningThreshold and criticalThreshold are durations before expiry.
func CheckExpiry(info vault.SecretInfo, warningThreshold, criticalThreshold time.Duration) SecretAlert {
	now := time.Now()
	timeLeft := info.ExpiresAt.Sub(now)

	alert := SecretAlert{
		Path:      info.Path,
		ExpiresAt: info.ExpiresAt,
		TimeLeft:  timeLeft,
	}

	switch {
	case timeLeft <= 0:
		alert.Status = StatusExpired
	case timeLeft <= criticalThreshold:
		alert.Status = StatusCritical
	case timeLeft <= warningThreshold:
		alert.Status = StatusWarning
	default:
		alert.Status = StatusOK
	}

	return alert
}

// CheckAll evaluates multiple secrets and returns alerts for non-OK statuses.
func CheckAll(secrets []vault.SecretInfo, warningThreshold, criticalThreshold time.Duration) []SecretAlert {
	var alerts []SecretAlert
	for _, s := range secrets {
		alert := CheckExpiry(s, warningThreshold, criticalThreshold)
		if alert.Status != StatusOK {
			alerts = append(alerts, alert)
		}
	}
	return alerts
}

// FormatAlert returns a formatted string describing the alert.
func FormatAlert(a SecretAlert) string {
	if a.Status == StatusExpired {
		return fmt.Sprintf("[%s] %s expired at %s", a.Status, a.Path, a.ExpiresAt.Format(time.RFC3339))
	}
	return fmt.Sprintf("[%s] %s expires in %s (at %s)",
		a.Status, a.Path,
		a.TimeLeft.Round(time.Second),
		a.ExpiresAt.Format(time.RFC3339),
	)
}
