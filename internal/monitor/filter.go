package monitor

import (
	"strings"
)

// FilterConfig defines rules for including or excluding secret paths from monitoring.
type FilterConfig struct {
	// IncludePrefixes, if non-empty, only paths matching at least one prefix are kept.
	IncludePrefixes []string
	// ExcludePrefixes removes paths matching any of these prefixes.
	ExcludePrefixes []string
}

// Filter applies include/exclude rules to a list of secret paths and returns
// the filtered subset. Include rules are applied first; exclude rules follow.
func Filter(paths []string, cfg FilterConfig) []string {
	result := make([]string, 0, len(paths))

	for _, p := range paths {
		if len(cfg.IncludePrefixes) > 0 && !matchesAny(p, cfg.IncludePrefixes) {
			continue
		}
		if matchesAny(p, cfg.ExcludePrefixes) {
			continue
		}
		result = append(result, p)
	}

	return result
}

// matchesAny returns true if s has any of the given prefixes.
func matchesAny(s string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}
