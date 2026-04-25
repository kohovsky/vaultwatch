package monitor_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

func TestParseInterval_Valid(t *testing.T) {
	d, err := monitor.ParseInterval("10m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 10*time.Minute {
		t.Errorf("expected 10m, got %s", d)
	}
}

func TestParseInterval_Empty_ReturnsDefault(t *testing.T) {
	d, err := monitor.ParseInterval("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 5*time.Minute {
		t.Errorf("expected default 5m, got %s", d)
	}
}

func TestParseInterval_Invalid(t *testing.T) {
	_, err := monitor.ParseInterval("not-a-duration")
	if err == nil {
		t.Fatal("expected error for invalid duration, got nil")
	}
}

func TestNextTick_IsFuture(t *testing.T) {
	before := time.Now()
	next := monitor.NextTick(1 * time.Minute)
	if !next.After(before) {
		t.Errorf("expected NextTick to be in the future, got %s", next)
	}
}

func TestNextTick_ApproximateInterval(t *testing.T) {
	interval := 30 * time.Second
	before := time.Now()
	next := monitor.NextTick(interval)
	diff := next.Sub(before)
	if diff < interval-time.Millisecond || diff > interval+time.Millisecond {
		t.Errorf("expected diff ~%s, got %s", interval, diff)
	}
}
