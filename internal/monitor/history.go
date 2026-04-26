package monitor

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// StatusRecord captures a snapshot of a secret's expiry status at a point in time.
type StatusRecord struct {
	Path      string    `json:"path"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
	RecordedAt time.Time `json:"recorded_at"`
}

// History tracks the last known status for each secret path.
type History struct {
	mu      sync.RWMutex
	records map[string]StatusRecord
	filePath string
}

// NewHistory creates a History, optionally loading persisted state from filePath.
func NewHistory(filePath string) (*History, error) {
	h := &History{
		records:  make(map[string]StatusRecord),
		filePath: filePath,
	}
	if filePath == "" {
		return h, nil
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return h, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, &h.records); err != nil {
		return nil, err
	}
	return h, nil
}

// HasChanged returns true if the status for path differs from the last recorded value.
func (h *History) HasChanged(path, status string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	prev, ok := h.records[path]
	if !ok {
		return true
	}
	return prev.Status != status
}

// Record stores the latest status for a path.
func (h *History) Record(path, status string, expiresAt time.Time) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.records[path] = StatusRecord{
		Path:       path,
		Status:     status,
		ExpiresAt:  expiresAt,
		RecordedAt: time.Now(),
	}
}

// Get returns the last recorded StatusRecord for the given path, and whether it exists.
func (h *History) Get(path string) (StatusRecord, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	record, ok := h.records[path]
	return record, ok
}

// Save persists the current history to disk (no-op if filePath is empty).
func (h *History) Save() error {
	if h.filePath == "" {
		return nil
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	data, err := json.MarshalIndent(h.records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(h.filePath, data, 0600)
}
