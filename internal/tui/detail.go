package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// viewDetail renders the detail pane for the currently selected session.
func (m Model) viewDetail() string {
	if len(m.filteredSessions) == 0 {
		return detailBorderStyle.Width(m.width - 4).Render(
			detailLabelStyle.Render("No session selected"),
		)
	}

	s := m.filteredSessions[m.selectedIndex]

	maxValWidth := m.width - 14 // label width + borders + padding
	if maxValWidth < 20 {
		maxValWidth = 20
	}

	// Tags line
	tagsStr := "-"
	if len(s.Tags) > 0 {
		tagsStr = formatTagsInline(s.Tags)
	}

	// First message
	firstMsg := s.FirstMessage
	if firstMsg == "" {
		firstMsg = "-"
	} else {
		firstMsg = fmt.Sprintf("%q", truncateStr(firstMsg, maxValWidth-4))
	}

	// Last message
	lastMsg := s.LastMessage
	if lastMsg == "" {
		lastMsg = "-"
	} else {
		lastMsg = fmt.Sprintf("%q", truncateStr(lastMsg, maxValWidth-4))
	}

	// CWD
	cwd := s.Cwd
	if cwd == "" {
		cwd = "-"
	}

	lines := []string{
		detailLabelStyle.Render("Tags:  ") + tagsStr,
		detailLabelStyle.Render("CWD:   ") + detailValueStyle.Render(truncateStr(cwd, maxValWidth)),
	}

	if s.Summary != "" {
		lines = append(lines, detailLabelStyle.Render("Sum:   ")+detailValueStyle.Render(truncateStr(s.Summary, maxValWidth)))
	}

	lines = append(lines,
		detailLabelStyle.Render("First: ")+detailValueStyle.Render(firstMsg),
		detailLabelStyle.Render("Last:  ")+detailValueStyle.Render(lastMsg),
	)

	content := strings.Join(lines, "\n")
	boxWidth := m.width - 4
	if boxWidth < 30 {
		boxWidth = 30
	}
	return detailBorderStyle.Width(boxWidth).Render(content)
}

// viewDeleteConfirm renders the inline delete confirmation prompt.
func (m Model) viewDeleteConfirm() string {
	if len(m.filteredSessions) == 0 {
		return ""
	}
	s := m.filteredSessions[m.selectedIndex]
	name := sessionDisplayName(s)
	prompt := deletePromptStyle.Render(fmt.Sprintf("Delete session %q? ", name)) +
		lipgloss.NewStyle().Foreground(colorMuted).Render("[y/n]")
	return prompt
}
