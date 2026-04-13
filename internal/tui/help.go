package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type keyBinding struct {
	key  string
	desc string
}

var helpBindings = []keyBinding{
	{"j / Down", "Move down"},
	{"k / Up", "Move up"},
	{"Enter", "Open selected session"},
	{"/", "Search (real-time filter)"},
	{"Esc", "Cancel search / Close dialog"},
	{"Tab", "Cycle project filter"},
	{"t", "Add/remove tag"},
	{"d", "Delete session"},
	{"?", "Toggle this help"},
	{"q", "Quit"},
	{"Ctrl+C", "Force quit"},
}

// viewHelp renders a full-screen help overlay with all key bindings.
func (m Model) viewHelp(width, height int) string {
	title := headerStyle.Render("Key Bindings")

	var lines []string
	lines = append(lines, title)
	lines = append(lines, "")

	for _, b := range helpBindings {
		line := helpKeyStyle.Render(b.key) + helpDescStyle.Render(b.desc)
		lines = append(lines, line)
	}

	lines = append(lines, "")
	lines = append(lines, footerStyle.Render("Press ? or Esc to close"))

	content := strings.Join(lines, "\n")

	boxWidth := 44
	if width > 0 && width < boxWidth+4 {
		boxWidth = width - 4
	}

	box := helpOverlayStyle.Width(boxWidth).Render(content)

	// Center the box
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
