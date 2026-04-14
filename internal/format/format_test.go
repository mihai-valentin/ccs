package format

import (
	"strings"
	"testing"
	"time"
)

func TestTruncate(t *testing.T) {
	cases := []struct {
		name   string
		in     string
		maxLen int
		want   string
	}{
		{"shorter than max", "hi", 10, "hi"},
		{"equal to max", "hello", 5, "hello"},
		{"longer ascii", "helloworld", 8, "hello..."},
		{"tiny max", "abcdef", 2, "ab"},
		{"max equal 3", "abcdef", 3, "abc"},
		{"multi-byte safe", "héllo wörld", 7, "héll..."},
		{"emoji safe", "hi 🐿️🐿️🐿️", 5, "hi..."},
		{"cjk runes", "你好世界朋友", 4, "你..."},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Truncate(tc.in, tc.maxLen)
			if got != tc.want {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tc.in, tc.maxLen, got, tc.want)
			}
		})
	}
}

func TestSessionDisplayName(t *testing.T) {
	cases := []struct {
		name, sessionName, id, want string
	}{
		{"prefers name", "my-session", "abcdef0123456789", "my-session"},
		{"falls back to 8-char id", "", "abcdef0123456789", "abcdef01"},
		{"short id passthrough", "", "abc", "abc"},
		{"empty both", "", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SessionDisplayName(tc.sessionName, tc.id)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFormatRelativeTime(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name    string
		t       time.Time
		wantSub string // substring expected in result
	}{
		{"zero → dash", time.Time{}, "-"},
		{"just now", now.Add(-30 * time.Second), "just now"},
		{"minutes", now.Add(-5 * time.Minute), "m ago"},
		{"hours", now.Add(-3 * time.Hour), "h ago"},
		{"days", now.Add(-5 * 24 * time.Hour), "d ago"},
		{"months", now.Add(-60 * 24 * time.Hour), "mo ago"},
		{"years", now.Add(-400 * 24 * time.Hour), "y ago"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatRelativeTime(tc.t)
			if !strings.Contains(got, tc.wantSub) {
				t.Errorf("FormatRelativeTime(%v) = %q, want substring %q", tc.t, got, tc.wantSub)
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	t.Run("RFC3339Nano", func(t *testing.T) {
		in := "2026-04-13T10:30:45.123456789Z"
		got := ParseTime(in)
		want := time.Date(2026, 4, 13, 10, 30, 45, 123456789, time.UTC)
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
	t.Run("RFC3339 fallback", func(t *testing.T) {
		got := ParseTime("2026-04-13T10:30:45Z")
		want := time.Date(2026, 4, 13, 10, 30, 45, 0, time.UTC)
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
	t.Run("invalid returns zero", func(t *testing.T) {
		got := ParseTime("not a time")
		if !got.IsZero() {
			t.Errorf("got %v, want zero time", got)
		}
	})
}
