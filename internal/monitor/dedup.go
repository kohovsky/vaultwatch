package monitor

import (
	"sync"
	"time"
)

// DedupWindow suppresses repeated alerts for the same path+status
// within a configurable cooldown window.
type DedupWindow struct {
	mu       sync.Mutex
	cooldown time.Duration
	seen     map[string]time.Time
}

// NewDedupWindow creates a DedupWindow with the given cooldown duration.
// If cooldown is zero, a default of 1 hour is used.
func NewDedupWindow(cooldown time.Duration) *DedupWindow {
	if cooldown <= 0 {
		cooldown = time.Hour
	}
	return &DedupWindow{
		cooldown: cooldown,
		seen:     make(map[string]time.Time),
	}
}

// IsDuplicate returns true if an alert for the given key was already
// seen within the cooldown window. It records the key if not duplicate.
func (d *DedupWindow) IsDuplicate(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if last, ok := d.seen[key]; ok {
		if time.Since(last) < d.cooldown {
			return true
		}
	}
	d.seen[key] = time.Now()
	return false
}

// Key builds a dedup key from a path and status string.
func DedupKey(path, status string) string {
	return path + "::" + status
}

// Evict removes expired entries to keep memory bounded.
func (d *DedupWindow) Evict() {
	d.mu.Lock()
	defer d.mu.Unlock()

	for k, t := range d.seen {
		if time.Since(t) >= d.cooldown {
			delete(d.seen, k)
		}
	}
}
