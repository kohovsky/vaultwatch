package monitor

import (
	"sync"
	"time"
)

// RateLimiter restricts how frequently alerts can be sent per path.
// It is safe for concurrent use.
type RateLimiter struct {
	mu       sync.Mutex
	lastSent map[string]time.Time
	minGap   time.Duration
}

// NewRateLimiter creates a RateLimiter that enforces a minimum gap between
// alerts for the same path.
func NewRateLimiter(minGap time.Duration) *RateLimiter {
	if minGap <= 0 {
		minGap = 5 * time.Minute
	}
	return &RateLimiter{
		lastSent: make(map[string]time.Time),
		minGap:   minGap,
	}
}

// Allow returns true if an alert for the given path is permitted at now.
// If allowed, it records now as the last-sent time for that path.
func (r *RateLimiter) Allow(path string, now time.Time) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	last, seen := r.lastSent[path]
	if seen && now.Sub(last) < r.minGap {
		return false
	}
	r.lastSent[path] = now
	return true
}

// Reset clears the rate-limit record for a specific path.
func (r *RateLimiter) Reset(path string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.lastSent, path)
}

// ResetAll clears all rate-limit records.
func (r *RateLimiter) ResetAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastSent = make(map[string]time.Time)
}
