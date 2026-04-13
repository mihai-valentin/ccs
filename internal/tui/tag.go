package tui

import (
	"fmt"
	"strings"
)

// viewTagDialog renders the inline tag prompt showing current tags
// and an input for adding/removing a tag.
func (m Model) viewTagDialog() string {
	if len(m.filteredSessions) == 0 || m.selectedIndex >= len(m.filteredSessions) {
		return ""
	}

	s := m.filteredSessions[m.selectedIndex]

	var lines []string

	// Show current tags
	if len(s.Tags) > 0 {
		var tagNames []string
		for _, t := range s.Tags {
			tagNames = append(tagNames, t.Name)
		}
		lines = append(lines,
			detailLabelStyle.Render("Current tags: ")+strings.Join(tagNames, ", "))
	} else {
		lines = append(lines, detailLabelStyle.Render("No tags"))
	}

	// Input prompt
	lines = append(lines,
		fmt.Sprintf("%s %s %s",
			detailLabelStyle.Render("Tag:"),
			m.tagInput.View(),
			footerStyle.Render("(enter=toggle, esc=cancel)"),
		))

	content := strings.Join(lines, "\n")
	return tagDialogStyle.Render(content)
}
