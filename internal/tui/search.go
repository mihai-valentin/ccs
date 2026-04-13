package tui

// Search component logic is integrated into app.go's updateSearch method
// and the searchInput textinput.Model field.
//
// The search input is activated with '/' in normal mode and filters the
// session list in real-time as the user types. Pressing Enter confirms
// the search, Esc clears it and returns to normal mode.
