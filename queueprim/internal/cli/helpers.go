package cli

import (
	"fmt"
	"time"
)

// parseDurationFlag parses a duration string that may include "d" for days
// (e.g. "7d" → 7*24h). Standard time.ParseDuration handles h/m/s/ms etc.
func parseDurationFlag(s string) (time.Duration, error) {
	// Handle simple "Nd" syntax for days since time.ParseDuration doesn't support it.
	var days int
	if n, err := fmt.Sscanf(s, "%dd", &days); n == 1 && err == nil {
		return time.Duration(days) * 24 * time.Hour, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q (use e.g. 30m, 2h, 7d)", s)
	}
	return d, nil
}
