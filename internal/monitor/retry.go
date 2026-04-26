package monitor

import (
	"errors"
	"time"
)

// RetryConfig holds configuration for retry behavior.
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

// DefaultRetryConfig returns a sensible default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   500 * time.Millisecond,
		MaxDelay:    10 * time.Second,
	}
}

// Retrier executes a function with exponential backoff retry logic.
type Retrier struct {
	cfg   RetryConfig
	sleep func(time.Duration)
}

// NewRetrier creates a Retrier with the given config.
func NewRetrier(cfg RetryConfig) *Retrier {
	return &Retrier{
		cfg:   cfg,
		sleep: time.Sleep,
	}
}

// Do executes fn up to MaxAttempts times, backing off exponentially between
// attempts. It returns the last error if all attempts fail.
func (r *Retrier) Do(fn func() error) error {
	var lastErr error
	delay := r.cfg.BaseDelay

	for attempt := 0; attempt < r.cfg.MaxAttempts; attempt++ {
		if err := fn(); err != nil {
			lastErr = err
			if attempt < r.cfg.MaxAttempts-1 {
				r.sleep(delay)
				delay *= 2
				if delay > r.cfg.MaxDelay {
					delay = r.cfg.MaxDelay
				}
			}
			continue
		}
		return nil
	}
	return lastErr
}

// ErrNonRetryable can be wrapped around an error to signal that retrying
// should be skipped entirely.
type ErrNonRetryable struct {
	Cause error
}

func (e *ErrNonRetryable) Error() string { return e.Cause.Error() }
func (e *ErrNonRetryable) Unwrap() error { return e.Cause }

// IsNonRetryable returns true if err is or wraps ErrNonRetryable.
func IsNonRetryable(err error) bool {
	var nr *ErrNonRetryable
	return errors.As(err, &nr)
}
