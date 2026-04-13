package cli

import (
	"fmt"
	"strings"

	"github.com/mihai-valentin/ccs/internal/model"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	var (
		all     bool
		tags    []string
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search sessions across name, messages, cwd, and branch",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.Join(args, " ")

			d, err := syncIndex()
			if err != nil {
				return err
			}
			defer d.Close()

			sessions, err := d.SearchSessions(query)
			if err != nil {
				return fmt.Errorf("searching sessions: %w", err)
			}

			// Filter by tags if specified
			if len(tags) > 0 {
				var filteredSessions []model.Session
				for _, s := range sessions {
					sTags, err := d.GetSessionTags(s.ID)
					if err != nil {
						continue
					}
					tagSet := make(map[string]bool)
					for _, t := range sTags {
						tagSet[t.Name] = true
					}
					match := true
					for _, want := range tags {
						if !tagSet[want] {
							match = false
							break
						}
					}
					if match {
						filteredSessions = append(filteredSessions, s)
					}
				}
				sessions = filteredSessions
			}

			// Filter by current project if not --all
			if !all {
				if projDir, err := detectProjectDir(); err == nil && projDir != "" {
					var filtered []model.Session
					for _, s := range sessions {
						if s.ProjectDir == projDir {
							filtered = append(filtered, s)
						}
					}
					sessions = filtered
				}
			}

			if jsonOut {
				return printJSON(sessions)
			}

			if len(sessions) == 0 {
				fmt.Println("No sessions found.")
				return nil
			}

			w := newTabWriter()
			fmt.Fprintln(w, "NAME\tPROJECT\tBRANCH\tMSGS\tUPDATED")
			for _, s := range sessions {
				name := sessionDisplayName(s.Name, s.ID)
				branch := s.GitBranch
				if branch == "" {
					branch = "-"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
					truncate(name, 30),
					truncate(s.ProjectDir, 20),
					truncate(branch, 15),
					s.MessageCount,
					formatRelativeTime(s.UpdatedAt),
				)
			}
			return w.Flush()
		},
	}

	cmd.Flags().BoolVarP(&all, "all", "a", false, "search all projects")
	cmd.Flags().StringSliceVarP(&tags, "tag", "t", nil, "filter by tag (repeatable)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "output as JSON")

	return cmd
}
