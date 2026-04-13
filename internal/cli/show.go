package cli

import (
	"fmt"
	"strings"

	"github.com/mihai/ccs/internal/opener"
	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "show <id|name>",
		Short: "Show full session details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := syncIndex()
			if err != nil {
				return err
			}
			defer d.Close()

			session, err := opener.ResolveSession(d, args[0])
			if err != nil {
				return err
			}

			// Attach tags
			tags, err := d.GetSessionTags(session.ID)
			if err != nil {
				return fmt.Errorf("fetching tags: %w", err)
			}
			session.Tags = tags

			if jsonOut {
				return printJSON(session)
			}

			fmt.Printf("Session:  %s\n", session.ID)
			if session.Name != "" {
				fmt.Printf("Name:     %s\n", session.Name)
			}
			fmt.Printf("Project:  %s\n", session.ProjectDir)
			fmt.Printf("CWD:      %s\n", session.Cwd)
			if session.GitBranch != "" {
				fmt.Printf("Branch:   %s\n", session.GitBranch)
			}
			fmt.Printf("Messages: %d\n", session.MessageCount)
			fmt.Printf("Created:  %s (%s)\n", session.CreatedAt.Format("2006-01-02 15:04"), formatRelativeTime(session.CreatedAt))
			fmt.Printf("Updated:  %s (%s)\n", session.UpdatedAt.Format("2006-01-02 15:04"), formatRelativeTime(session.UpdatedAt))

			if len(tags) > 0 {
				names := make([]string, len(tags))
				for i, t := range tags {
					names[i] = t.Name
				}
				fmt.Printf("Tags:     %s\n", strings.Join(names, ", "))
			}

			if session.Summary != "" {
				fmt.Printf("\nSummary:\n  %s\n", session.Summary)
			}

			if session.FirstMessage != "" {
				fmt.Printf("\nFirst message:\n  %s\n", session.FirstMessage)
			}
			if session.LastMessage != "" {
				fmt.Printf("\nLast message:\n  %s\n", session.LastMessage)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "output as JSON")

	return cmd
}
