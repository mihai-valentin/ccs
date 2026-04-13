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
		lines = append(lines, detailLabelStyle.Render("Summary: ")+detailValueStyle.Render(truncateStr(s.Summary, maxValWidth)))
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

// viewSummaryOverlay renders the full summary as a centered overlay.
func (m Model) viewSummaryOverlay() string {
	if len(m.filteredSessions) == 0 {
		return ""
	}
	s := m.filteredSessions[m.selectedIndex]
	name := sessionDisplayName(s)

	maxWidth := m.width - 8
	if maxWidth < 40 {
		maxWidth = 40
	}
	if maxWidth > 100 {
		maxWidth = 100
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorPrimary).
		Render("Summary: " + name)

	body := s.Summary
	if body == "" {
		body = "No summary available."
	}

	// Word-wrap the summary text
	body = wordWrap(body, maxWidth-4)

	footer := lipgloss.NewStyle().
		Foreground(colorMuted).
		Render("\n[s/Esc] Close")

	content := title + "\n\n" + body + footer

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(1, 2).
		Width(maxWidth).
		Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// wordWrap wraps text to the given width at word boundaries.
func wordWrap(text string, width int) string {
	if width <= 0 {
		return text
	}
	var lines []string
	for _, paragraph := range strings.Split(text, "\n") {
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}
		line := words[0]
		for _, w := range words[1:] {
			if len(line)+1+len(w) > width {
				lines = append(lines, line)
				line = w
			} else {
				line += " " + w
			}
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
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
