package monitor

import "time"

// ParseInterval parses a duration string and returns the interval.
// It returns a default of 5 minutes if the string is empty.
func ParseInterval(s string) (time.Duration, error) {
	if s == "" {
		return 5 * time.Minute, nil
	}
	return time.ParseDuration(s)
}

// NextTick returns the next tick time from now given an interval.
func NextTick(interval time.Duration) time.Time {
	return time.Now().Add(interval)
}
