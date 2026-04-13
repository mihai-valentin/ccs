package format

import (
	"fmt"
	"time"
)

// FormatRelativeTime formats a time as a human-readable relative duration.
func FormatRelativeTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	case d < 365*24*time.Hour:
		return fmt.Sprintf("%dmo ago", int(d.Hours()/24/30))
	default:
		return fmt.Sprintf("%dy ago", int(d.Hours()/24/365))
	}
}

// SessionDisplayName returns the session's name if set, otherwise the first 8
// characters of its ID.
func SessionDisplayName(name string, id string) string {
	if name != "" {
		return name
	}
	if len(id) >= 8 {
		return id[:8]
	}
	return id
}

// Truncate returns s truncated to maxLen runes with a "..." suffix when
// truncation occurs. It operates on runes rather than bytes to avoid splitting
// multi-byte characters.
func Truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}
