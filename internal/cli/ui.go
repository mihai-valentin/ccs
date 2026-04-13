package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mihai/ccs/internal/model"
	"github.com/mihai/ccs/internal/opener"
	"github.com/mihai/ccs/internal/tui"
	"github.com/spf13/cobra"
)

func newUICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ui",
		Short: "Launch interactive TUI",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := syncIndex()
			if err != nil {
				return err
			}
			defer d.Close()

			// Load all sessions (no limit for TUI)
			sessions, err := d.ListSessions(model.SessionFilter{
				Limit:  -1,
				SortBy: "updated",
			})
			if err != nil {
				return fmt.Errorf("loading sessions: %w", err)
			}

			m := tui.NewModel(sessions, d, getClaudeDir())

			p := tea.NewProgram(m, tea.WithAltScreen())
			finalModel, err := p.Run()
			if err != nil {
				return fmt.Errorf("TUI error: %w", err)
			}

			// Check if the user selected a session to open
			if fm, ok := finalModel.(tui.Model); ok && fm.SessionToOpen != nil {
				return opener.OpenSession(*fm.SessionToOpen, false)
			}

			return nil
		},
	}
}
