package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mihai/ccs/internal/db"
	"github.com/mihai/ccs/internal/model"
)

// Mode represents the current interaction mode of the TUI.
type Mode int

const (
	ModeNormal Mode = iota
	ModeSearch
	ModeTag
	ModeDeleteConfirm
	ModeHelp
	ModeSummary
)

// Model is the main bubbletea model for the TUI.
type Model struct {
	// Data
	allSessions      []model.Session // unfiltered full list
	filteredSessions []model.Session // after search + project filter
	projects         []string        // discovered project names

	// Navigation
	selectedIndex int
	scrollOffset  int // top visible row index

	// Filters
	searchQuery    string
	projectFilter  string // "" means "All"
	projectIndex   int    // index into projects list; 0 = All

	// UI state
	mode       Mode
	width      int
	height     int
	statusMsg  string
	statusTime time.Time

	// Components
	searchInput textinput.Model
	tagInput    textinput.Model

	// Database handle (for tag operations)
	db *db.DB

	// Result: session to open after TUI exits
	SessionToOpen *model.Session
}

// NewModel creates a new TUI model with the given sessions and database handle.
func NewModel(sessions []model.Session, database *db.DB) Model {
	si := textinput.New()
	si.Placeholder = "search..."
	si.CharLimit = 100

	ti := textinput.New()
	ti.Placeholder = "tag name"
	ti.CharLimit = 50

	// Extract unique projects
	projectSet := make(map[string]bool)
	for i := range sessions {
		projectSet[sessions[i].ProjectDir] = true
		// Load tags for each session
		if database != nil {
			if tags, err := database.GetSessionTags(sessions[i].ID); err == nil {
				sessions[i].Tags = tags
			}
		}
	}
	var projects []string
	for p := range projectSet {
		projects = append(projects, p)
	}

	m := Model{
		allSessions: sessions,
		projects:    projects,
		searchInput: si,
		tagInput:    ti,
		db:          database,
	}
	m.applyFilters()
	return m
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.WindowSize()
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Global keys that work in any mode
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

		switch m.mode {
		case ModeNormal:
			return m.updateNormal(msg)
		case ModeSearch:
			return m.updateSearch(msg)
		case ModeTag:
			return m.updateTag(msg)
		case ModeDeleteConfirm:
			return m.updateDeleteConfirm(msg)
		case ModeHelp:
			return m.updateHelp(msg)
		case ModeSummary:
			return m.updateSummary(msg)
		}
	}
	return m, nil
}

func (m Model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		return m, tea.Quit
	case "j", "down":
		m.moveDown()
	case "k", "up":
		m.moveUp()
	case "enter":
		if len(m.filteredSessions) > 0 {
			s := m.filteredSessions[m.selectedIndex]
			m.SessionToOpen = &s
			return m, tea.Quit
		}
	case "/":
		m.mode = ModeSearch
		m.searchInput.SetValue(m.searchQuery)
		m.searchInput.Focus()
		return m, m.searchInput.Focus()
	case "t":
		if len(m.filteredSessions) > 0 {
			m.mode = ModeTag
			m.tagInput.SetValue("")
			m.tagInput.Focus()
			return m, m.tagInput.Focus()
		}
	case "d":
		if len(m.filteredSessions) > 0 {
			m.mode = ModeDeleteConfirm
		}
	case "tab":
		m.cycleProjectFilter()
	case "s":
		if len(m.filteredSessions) > 0 && m.filteredSessions[m.selectedIndex].Summary != "" {
			m.mode = ModeSummary
		}
	case "?":
		m.mode = ModeHelp
	}
	return m, nil
}

func (m Model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.searchQuery = ""
		m.searchInput.SetValue("")
		m.searchInput.Blur()
		m.applyFilters()
		return m, nil
	case "enter":
		m.mode = ModeNormal
		m.searchQuery = m.searchInput.Value()
		m.searchInput.Blur()
		m.applyFilters()
		return m, nil
	default:
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		m.searchQuery = m.searchInput.Value()
		m.applyFilters()
		return m, cmd
	}
}

func (m Model) updateTag(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.tagInput.Blur()
		return m, nil
	case "enter":
		tagName := strings.TrimSpace(m.tagInput.Value())
		if tagName != "" && m.db != nil && len(m.filteredSessions) > 0 {
			session := &m.filteredSessions[m.selectedIndex]
			// Toggle: if tag exists, remove it; otherwise add it
			hasTag := false
			for _, t := range session.Tags {
				if t.Name == tagName {
					hasTag = true
					break
				}
			}
			if hasTag {
				if err := m.db.RemoveTag(session.ID, tagName); err != nil {
					m.statusMsg = "Error removing tag: " + err.Error()
				} else {
					m.statusMsg = "Removed tag: " + tagName
				}
			} else {
				if err := m.db.AddTag(session.ID, tagName); err != nil {
					m.statusMsg = "Error adding tag: " + err.Error()
				} else {
					m.statusMsg = "Added tag: " + tagName
				}
			}
			m.statusTime = time.Now()
			// Refresh tags for all sessions
			m.refreshTags()
		}
		m.mode = ModeNormal
		m.tagInput.Blur()
		return m, nil
	default:
		var cmd tea.Cmd
		m.tagInput, cmd = m.tagInput.Update(msg)
		return m, cmd
	}
}

func (m Model) updateDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if m.db != nil && len(m.filteredSessions) > 0 {
			session := m.filteredSessions[m.selectedIndex]
			if err := m.db.DeleteSession(session.ID); err != nil {
				m.statusMsg = "Error deleting session: " + err.Error()
			} else {
				m.statusMsg = "Deleted session: " + sessionDisplayName(session)
				// Remove from allSessions
				for i, s := range m.allSessions {
					if s.ID == session.ID {
						m.allSessions = append(m.allSessions[:i], m.allSessions[i+1:]...)
						break
					}
				}
				m.applyFilters()
				if m.selectedIndex >= len(m.filteredSessions) && m.selectedIndex > 0 {
					m.selectedIndex--
				}
			}
			m.statusTime = time.Now()
		}
		m.mode = ModeNormal
		return m, nil
	case "n", "N", "esc":
		m.mode = ModeNormal
		return m, nil
	}
	return m, nil
}

func (m Model) updateSummary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "s", "esc", "q":
		m.mode = ModeNormal
	}
	return m, nil
}

func (m Model) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "?", "esc", "q":
		m.mode = ModeNormal
	}
	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	if m.mode == ModeHelp {
		return m.viewHelp(m.width, m.height)
	}

	if m.mode == ModeSummary {
		return m.viewSummaryOverlay()
	}

	var sections []string

	// Top bar: search + filter
	sections = append(sections, m.viewTopBar())

	// Session list
	listHeight := m.listHeight()
	sections = append(sections, m.viewList(listHeight))

	// Detail pane
	sections = append(sections, m.viewDetail())

	// Status message (if recent)
	if m.statusMsg != "" && time.Since(m.statusTime) < 5*time.Second {
		sections = append(sections, successStyle.Render(m.statusMsg))
	}

	// Tag dialog or delete confirm (overlay above footer)
	switch m.mode {
	case ModeTag:
		sections = append(sections, m.viewTagDialog())
	case ModeDeleteConfirm:
		sections = append(sections, m.viewDeleteConfirm())
	}

	// Footer
	sections = append(sections, m.viewFooter())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) viewTopBar() string {
	var parts []string

	title := headerStyle.Render("ccs")
	parts = append(parts, title)

	if m.mode == ModeSearch {
		label := searchLabelStyle.Render("Search: ")
		parts = append(parts, label+m.searchInput.View())
	} else if m.searchQuery != "" {
		parts = append(parts, searchLabelStyle.Render("Search: ")+m.searchQuery)
	}

	filterLabel := "All"
	if m.projectFilter != "" {
		filterLabel = m.projectFilter
	}
	parts = append(parts, filterStyle.Render("["+filterLabel+"]"))

	return lipgloss.JoinHorizontal(lipgloss.Center, strings.Join(parts, "  "))
}

func (m Model) viewFooter() string {
	keys := []struct{ key, desc string }{
		{"Enter", "Open"},
		{"s", "Summary"},
		{"t", "Tag"},
		{"d", "Delete"},
		{"/", "Search"},
		{"Tab", "Filter"},
		{"?", "Help"},
		{"q", "Quit"},
	}
	var parts []string
	for _, k := range keys {
		parts = append(parts, footerKeyStyle.Render("["+k.key+"]")+" "+footerStyle.Render(k.desc))
	}
	return footerStyle.Render(strings.Join(parts, "  "))
}

func (m *Model) moveDown() {
	if m.selectedIndex < len(m.filteredSessions)-1 {
		m.selectedIndex++
		// Scroll if needed
		visibleRows := m.listHeight()
		if m.selectedIndex >= m.scrollOffset+visibleRows {
			m.scrollOffset = m.selectedIndex - visibleRows + 1
		}
	}
}

func (m *Model) moveUp() {
	if m.selectedIndex > 0 {
		m.selectedIndex--
		if m.selectedIndex < m.scrollOffset {
			m.scrollOffset = m.selectedIndex
		}
	}
}

func (m *Model) cycleProjectFilter() {
	m.projectIndex++
	if m.projectIndex > len(m.projects) {
		m.projectIndex = 0
	}
	if m.projectIndex == 0 {
		m.projectFilter = ""
	} else {
		m.projectFilter = m.projects[m.projectIndex-1]
	}
	m.applyFilters()
}

func (m *Model) applyFilters() {
	m.filteredSessions = nil
	query := strings.ToLower(m.searchQuery)
	for _, s := range m.allSessions {
		// Project filter
		if m.projectFilter != "" && s.ProjectDir != m.projectFilter {
			continue
		}
		// Search filter
		if query != "" {
			match := strings.Contains(strings.ToLower(s.Name), query) ||
				strings.Contains(strings.ToLower(s.ID), query) ||
				strings.Contains(strings.ToLower(s.Cwd), query) ||
				strings.Contains(strings.ToLower(s.GitBranch), query) ||
				strings.Contains(strings.ToLower(s.FirstMessage), query) ||
				strings.Contains(strings.ToLower(s.LastMessage), query)
			if !match {
				continue
			}
		}
		m.filteredSessions = append(m.filteredSessions, s)
	}
	// Reset selection if out of bounds
	if m.selectedIndex >= len(m.filteredSessions) {
		m.selectedIndex = 0
		m.scrollOffset = 0
	}
}

func (m *Model) refreshTags() {
	if m.db == nil {
		return
	}
	for i := range m.allSessions {
		if tags, err := m.db.GetSessionTags(m.allSessions[i].ID); err == nil {
			m.allSessions[i].Tags = tags
		}
	}
	m.applyFilters()
}

func (m Model) listHeight() int {
	// Reserve: top bar (1) + table header (2) + detail pane (6) + footer (1) + status (1) + dialog (2)
	h := m.height - 13
	if h < 3 {
		h = 3
	}
	return h
}

func sessionDisplayName(s model.Session) string {
	if s.Name != "" {
		return s.Name
	}
	if len(s.ID) >= 8 {
		return s.ID[:8]
	}
	return s.ID
}

func isStale(s model.Session) bool {
	return time.Since(s.UpdatedAt) > 30*24*time.Hour
}
