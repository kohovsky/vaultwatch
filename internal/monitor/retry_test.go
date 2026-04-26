package monitor

import (
	"errors"
	"testing"
	"time"
)

func TestRetrier_SuccessOnFirstAttempt(t *testing.T) {
	r := NewRetrier(DefaultRetryConfig())
	r.sleep = func(time.Duration) {}

	calls := 0
	err := r.Do(func() error {
		calls++
		return nil
	})

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestRetrier_RetriesOnError(t *testing.T) {
	cfg := RetryConfig{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: time.Second}
	r := NewRetrier(cfg)
	r.sleep = func(time.Duration) {}

	calls := 0
	sentinel := errors.New("transient")
	err := r.Do(func() error {
		calls++
		return sentinel
	})

	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestRetrier_SucceedsOnSecondAttempt(t *testing.T) {
	cfg := RetryConfig{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: time.Second}
	r := NewRetrier(cfg)
	r.sleep = func(time.Duration) {}

	calls := 0
	err := r.Do(func() error {
		calls++
		if calls < 2 {
			return errors.New("not yet")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}

func TestRetrier_ExponentialBackoff(t *testing.T) {
	cfg := RetryConfig{MaxAttempts: 4, BaseDelay: 100 * time.Millisecond, MaxDelay: 500 * time.Millisecond}
	r := NewRetrier(cfg)

	var sleeps []time.Duration
	r.sleep = func(d time.Duration) { sleeps = append(sleeps, d) }

	_ = r.Do(func() error { return errors.New("fail") })

	expected := []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 400 * time.Millisecond}
	if len(sleeps) != len(expected) {
		t.Fatalf("expected %d sleeps, got %d", len(expected), len(sleeps))
	}
	for i, d := range expected {
		if sleeps[i] != d {
			t.Errorf("sleep[%d]: expected %v, got %v", i, d, sleeps[i])
		}
	}
}

func TestRetrier_MaxDelayCapped(t *testing.T) {
	cfg := RetryConfig{MaxAttempts: 5, BaseDelay: 200 * time.Millisecond, MaxDelay: 300 * time.Millisecond}
	r := NewRetrier(cfg)

	var sleeps []time.Duration
	r.sleep = func(d time.Duration) { sleeps = append(sleeps, d) }

	_ = r.Do(func() error { return errors.New("fail") })

	for _, d := range sleeps {
		if d > cfg.MaxDelay {
			t.Errorf("sleep %v exceeded max delay %v", d, cfg.MaxDelay)
		}
	}
}

func TestIsNonRetryable(t *testing.T) {
	cause := errors.New("permanent failure")
	wrapped := &ErrNonRetryable{Cause: cause}

	if !IsNonRetryable(wrapped) {
		t.Error("expected IsNonRetryable to return true")
	}
	if !errors.Is(wrapped, cause) {
		t.Error("expected wrapped to unwrap to cause")
	}
	if IsNonRetryable(cause) {
		t.Error("expected plain error to not be non-retryable")
	}
}
