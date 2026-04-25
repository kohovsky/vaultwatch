package monitor

import (
	"testing"
	"time"

	"github.com/user/vaultwatch/internal/vault"
)

// TestCheckAll_RecordsHistory verifies that CheckAll updates History entries
// and that HasChanged reflects status transitions correctly.
func TestCheckAll_RecordsHistory(t *testing.T) {
	now := time.Now()
	secrets := []vault.SecretInfo{
		{Path: "secret/alpha", ExpiresAt: now.Add(10 * 24 * time.Hour)}, // ok
		{Path: "secret/beta", ExpiresAt: now.Add(12 * time.Hour)},        // warning
	}
	thresholds := []time.Duration{
		7 * 24 * time.Hour,  // warning
		24 * time.Hour,      // critical
	}

	h, err := NewHistory("")
	if err != nil {
		t.Fatalf("NewHistory: %v", err)
	}

	statuses := CheckAll(secrets, thresholds)

	for _, s := range statuses {
		if h.HasChanged(s.Path, string(s.Level)) {
			h.Record(s.Path, string(s.Level), s.ExpiresAt)
		}
	}

	// Second pass — same data, no changes expected
	for _, s := range statuses {
		if h.HasChanged(s.Path, string(s.Level)) {
			t.Errorf("path %q should not have changed on second pass", s.Path)
		}
	}

	// Simulate status change for beta
	h.Record("secret/beta", "critical", now.Add(20*time.Hour))
	if !h.HasChanged("secret/beta", "warning") {
		t.Error("expected HasChanged=true after status transition")
	}
}
