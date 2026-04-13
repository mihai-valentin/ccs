package tui

// Search component logic is integrated into app.go's updateSearch method
// and the searchInput textinput.Model field.
//
// The search input is activated with '/' in normal mode and filters the
// session list in real-time as the user types. Pressing Enter confirms
// the search, Esc clears it and returns to normal mode.
//
// This file provides any additional search-related helpers.

import "strings"

// matchesSearch checks if a session matches the given search query
// by searching across name, ID, CWD, branch, and messages.
func matchesSearch(query string, fields ...string) bool {
	q := strings.ToLower(query)
	for _, f := range fields {
		if strings.Contains(strings.ToLower(f), q) {
			return true
		}
	}
	return false
}
