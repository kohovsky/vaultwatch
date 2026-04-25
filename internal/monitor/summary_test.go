package monitor

import (
	"strings"
	"testing"
	"time"
)

func makeStatusEntry(path, status string, ttl time.Duration) SecretStatus {
	return SecretStatus{
		Path:      path,
		Status:    status,
		TTL:       ttl,
		ExpiresAt: time.Now().Add(ttl),
	}
}

func TestSummarize_Counts(t *testing.T) {
	statuses := []SecretStatus{
		makeStatusEntry("secret/a", StatusOK, 48*time.Hour),
		makeStatusEntry("secret/b", StatusWarning, 6*time.Hour),
		makeStatusEntry("secret/c", StatusCritical, 1*time.Hour),
		makeStatusEntry("secret/d", StatusExpired, -1*time.Minute),
		makeStatusEntry("secret/e", StatusOK, 72*time.Hour),
	}

	s := Summarize(statuses)

	if s.Total != 5 {
		t.Errorf("expected Total=5, got %d", s.Total)
	}
	if s.OK != 2 {
		t.Errorf("expected OK=2, got %d", s.OK)
	}
	if s.Warning != 1 {
		t.Errorf("expected Warning=1, got %d", s.Warning)
	}
	if s.Critical != 1 {
		t.Errorf("expected Critical=1, got %d", s.Critical)
	}
	if s.Expired != 1 {
		t.Errorf("expected Expired=1, got %d", s.Expired)
	}
}

func TestSummarize_Empty(t *testing.T) {
	s := Summarize(nil)
	if s.Total != 0 || s.OK != 0 {
		t.Errorf("expected all zeros for empty input, got %+v", s)
	}
}

func TestFormatSummary_ContainsFields(t *testing.T) {
	s := StatusSummary{Total: 3, OK: 1, Warning: 1, Critical: 1, AsOf: time.Now()}
	out := FormatSummary(s)
	for _, want := range []string{"Total", "OK", "Warning", "Critical", "Expired", "VaultWatch"} {
		if !strings.Contains(out, want) {
			t.Errorf("FormatSummary output missing %q", want)
		}
	}
}

func TestSortedStatuses_Order(t *testing.T) {
	statuses := []SecretStatus{
		makeStatusEntry("secret/ok", StatusOK, 48*time.Hour),
		makeStatusEntry("secret/warn", StatusWarning, 6*time.Hour),
		makeStatusEntry("secret/exp", StatusExpired, -1*time.Minute),
		makeStatusEntry("secret/crit", StatusCritical, 1*time.Hour),
	}

	sorted := SortedStatuses(statuses)

	expectedOrder := []string{StatusExpired, StatusCritical, StatusWarning, StatusOK}
	for i, st := range sorted {
		if st.Status != expectedOrder[i] {
			t.Errorf("position %d: expected status %q, got %q", i, expectedOrder[i], st.Status)
		}
	}
}

func TestSortedStatuses_DoesNotMutateOriginal(t *testing.T) {
	statuses := []SecretStatus{
		makeStatusEntry("secret/a", StatusOK, 48*time.Hour),
		makeStatusEntry("secret/b", StatusExpired, -1*time.Minute),
	}
	origFirst := statuses[0].Path
	_ = SortedStatuses(statuses)
	if statuses[0].Path != origFirst {
		t.Error("SortedStatuses mutated the original slice")
	}
}
