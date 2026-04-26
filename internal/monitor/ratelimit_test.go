package monitor

import (
	"testing"
	"time"
)

func TestRateLimiter_FirstCallAllowed(t *testing.T) {
	rl := NewRateLimiter(10 * time.Minute)
	now := time.Now()
	if !rl.Allow("secret/foo", now) {
		t.Fatal("expected first call to be allowed")
	}
}

func TestRateLimiter_SecondCallBlocked(t *testing.T) {
	rl := NewRateLimiter(10 * time.Minute)
	now := time.Now()
	rl.Allow("secret/foo", now)
	if rl.Allow("secret/foo", now.Add(1*time.Minute)) {
		t.Fatal("expected second call within gap to be blocked")
	}
}

func TestRateLimiter_AllowedAfterGap(t *testing.T) {
	rl := NewRateLimiter(5 * time.Minute)
	now := time.Now()
	rl.Allow("secret/foo", now)
	if !rl.Allow("secret/foo", now.Add(6*time.Minute)) {
		t.Fatal("expected call after gap to be allowed")
	}
}

func TestRateLimiter_DifferentPathsIndependent(t *testing.T) {
	rl := NewRateLimiter(10 * time.Minute)
	now := time.Now()
	rl.Allow("secret/foo", now)
	if !rl.Allow("secret/bar", now) {
		t.Fatal("expected different path to be allowed immediately")
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	rl := NewRateLimiter(10 * time.Minute)
	now := time.Now()
	rl.Allow("secret/foo", now)
	rl.Reset("secret/foo")
	if !rl.Allow("secret/foo", now.Add(1*time.Second)) {
		t.Fatal("expected call to be allowed after reset")
	}
}

func TestRateLimiter_ResetAll(t *testing.T) {
	rl := NewRateLimiter(10 * time.Minute)
	now := time.Now()
	rl.Allow("secret/foo", now)
	rl.Allow("secret/bar", now)
	rl.ResetAll()
	if !rl.Allow("secret/foo", now.Add(1*time.Second)) {
		t.Fatal("expected foo to be allowed after ResetAll")
	}
	if !rl.Allow("secret/bar", now.Add(1*time.Second)) {
		t.Fatal("expected bar to be allowed after ResetAll")
	}
}

func TestNewRateLimiter_DefaultsOnZeroGap(t *testing.T) {
	rl := NewRateLimiter(0)
	if rl.minGap != 5*time.Minute {
		t.Fatalf("expected default gap of 5m, got %v", rl.minGap)
	}
}
