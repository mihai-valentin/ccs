package format

import "time"

// ParseTime parses a timestamp string, trying RFC3339Nano first and falling
// back to RFC3339. Returns the zero time if neither format matches.
func ParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		t, _ = time.Parse(time.RFC3339, s)
	}
	return t
}
