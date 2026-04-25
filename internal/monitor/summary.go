package monitor

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// StatusSummary holds aggregated statistics across all monitored secrets.
type StatusSummary struct {
	Total    int
	OK       int
	Warning  int
	Critical int
	Expired  int
	AsOf     time.Time
}

// Summarize aggregates a slice of SecretStatus into a StatusSummary.
func Summarize(statuses []SecretStatus) StatusSummary {
	s := StatusSummary{AsOf: time.Now(), Total: len(statuses)}
	for _, st := range statuses {
		switch st.Status {
		case StatusOK:
			s.OK++
		case StatusWarning:
			s.Warning++
		case StatusCritical:
			s.Critical++
		case StatusExpired:
			s.Expired++
		}
	}
	return s
}

// FormatSummary returns a human-readable summary string.
func FormatSummary(s StatusSummary) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "VaultWatch Summary [%s]\n", s.AsOf.Format(time.RFC3339))
	fmt.Fprintf(&sb, "  Total:    %d\n", s.Total)
	fmt.Fprintf(&sb, "  OK:       %d\n", s.OK)
	fmt.Fprintf(&sb, "  Warning:  %d\n", s.Warning)
	fmt.Fprintf(&sb, "  Critical: %d\n", s.Critical)
	fmt.Fprintf(&sb, "  Expired:  %d\n", s.Expired)
	return sb.String()
}

// SortedStatuses returns a copy of statuses sorted by severity (expired first)
// and then by path alphabetically within the same severity.
func SortedStatuses(statuses []SecretStatus) []SecretStatus {
	copy_ := make([]SecretStatus, len(statuses))
	copy(copy_, statuses)
	order := map[string]int{
		StatusExpired:  0,
		StatusCritical: 1,
		StatusWarning:  2,
		StatusOK:       3,
	}
	sort.SliceStable(copy_, func(i, j int) bool {
		oi := order[copy_[i].Status]
		oj := order[copy_[j].Status]
		if oi != oj {
			return oi < oj
		}
		return copy_[i].Path < copy_[j].Path
	})
	return copy_
}
