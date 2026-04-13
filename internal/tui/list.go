package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mihai/ccs/internal/format"
	"github.com/mihai/ccs/internal/model"
)

// viewList renders the session list table with scrolling.
func (m Model) viewList(visibleRows int) string {
	if len(m.filteredSessions) == 0 {
		return lipgloss.NewStyle().
			Foreground(colorMuted).
			Padding(1, 2).
			Render("No sessions found.")
	}

	// Column widths (adaptive to terminal width)
	colNum := 4
	colName := 22
	colProject := 18
	colBranch := 14
	colUpdated := 10

	totalCols := colNum + colName + colProject + colBranch + colUpdated + 8 // separators
	if m.width > totalCols+20 {
		extra := m.width - totalCols - 4
		colName += extra / 2
		colProject += extra / 4
		colBranch += extra / 4
	}

	// Header
	header := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s",
		colNum, "#",
		colName, "Name/ID",
		colProject, "Project",
		colBranch, "Branch",
		colUpdated, "Updated",
	)
	lines := []string{tableHeaderStyle.Render(header)}

	// Visible rows
	end := m.scrollOffset + visibleRows
	if end > len(m.filteredSessions) {
		end = len(m.filteredSessions)
	}

	for i := m.scrollOffset; i < end; i++ {
		s := m.filteredSessions[i]
		lines = append(lines, m.renderRow(i, s, colNum, colName, colProject, colBranch, colUpdated))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderRow(idx int, s model.Session, colNum, colName, colProject, colBranch, colUpdated int) string {
	selected := idx == m.selectedIndex

	marker := "  "
	if selected {
		marker = "> "
	}

	name := sessionDisplayName(s)
	if len(name) > colName-1 {
		name = name[:colName-4] + "..."
	}

	branch := s.GitBranch
	if branch == "" {
		branch = "-"
	}
	if len(branch) > colBranch-1 {
		branch = branch[:colBranch-4] + "..."
	}

	project := s.ProjectDir
	if len(project) > colProject-1 {
		project = project[:colProject-4] + "..."
	}

	updated := formatRelativeTime(s.UpdatedAt)

	row := fmt.Sprintf("%s%-*d %-*s %-*s %-*s %-*s",
		marker,
		colNum-2, idx+1,
		colName, name,
		colProject, project,
		colBranch, branch,
		colUpdated, updated,
	)

	stale := isStale(s)
	switch {
	case selected && stale:
		return staleSelectedRowStyle.Render(row)
	case selected:
		return selectedRowStyle.Render(row)
	case stale:
		return staleRowStyle.Render(row)
	default:
		return normalRowStyle.Render(row)
	}
}

func formatRelativeTime(t time.Time) string {
	return format.FormatRelativeTime(t)
}

func truncateStr(s string, maxLen int) string {
	return format.Truncate(s, maxLen)
}

func formatTagsInline(tags []model.Tag) string {
	if len(tags) == 0 {
		return ""
	}
	var parts []string
	for _, t := range tags {
		parts = append(parts, tagStyle.Render(t.Name))
	}
	return strings.Join(parts, " ")
}
