package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	colorPrimary   = lipgloss.Color("#7C3AED") // purple
	colorSecondary = lipgloss.Color("#06B6D4") // cyan
	colorMuted     = lipgloss.Color("#6B7280") // gray
	colorStale     = lipgloss.Color("#9CA3AF") // light gray for >30d old
	colorTagBg     = lipgloss.Color("#4C1D95") // dark purple
	colorTagFg     = lipgloss.Color("#DDD6FE") // light purple
	colorSelected  = lipgloss.Color("#7C3AED")
	colorBorder    = lipgloss.Color("#4B5563")
	colorError     = lipgloss.Color("#EF4444")
	colorSuccess   = lipgloss.Color("#10B981")

	// Header / top bar
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			Padding(0, 1)

	// Session list table
	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorSecondary).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(colorBorder)

	selectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(colorSelected)

	normalRowStyle = lipgloss.NewStyle()

	staleRowStyle = lipgloss.NewStyle().
			Foreground(colorStale)

	staleSelectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorStale).
				Background(colorSelected)

	// Detail pane
	detailBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorBorder).
				Padding(0, 1)

	detailLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorSecondary)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E5E7EB"))

	// Tags
	tagStyle = lipgloss.NewStyle().
			Foreground(colorTagFg).
			Background(colorTagBg).
			Padding(0, 1)

	// Footer / help
	footerStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	footerKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#E5E7EB"))

	// Search
	searchLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorSecondary)

	// Filter indicator
	filterStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	// Tag dialog
	tagDialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 1)

	// Help overlay
	helpOverlayStyle = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(colorPrimary).
				Padding(1, 2)

	helpKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorSecondary).
			Width(12)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5E7EB"))

	// Error/status
	errorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)

	// Delete confirmation
	deletePromptStyle = lipgloss.NewStyle().
				Foreground(colorError).
				Bold(true)
)
