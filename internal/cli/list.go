package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/mihai/ccs/internal/model"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var (
		all     bool
		project string
		tags    []string
		limit   int
		sortBy  string
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := syncIndex()
			if err != nil {
				return err
			}
			defer d.Close()

			f := model.SessionFilter{
				Tags:   tags,
				Limit:  limit,
				SortBy: sortBy,
			}

			if !all && project == "" {
				// Default to current project directory name
				if cwd, err := detectProjectDir(); err == nil && cwd != "" {
					f.ProjectDir = cwd
				}
			}
			if project != "" {
				f.ProjectDir = project
			}

			sessions, err := d.ListSessions(f)
			if err != nil {
				return fmt.Errorf("listing sessions: %w", err)
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
				tagStr := ""
				if sTags, err := d.GetSessionTags(s.ID); err == nil && len(sTags) > 0 {
					names := make([]string, len(sTags))
					for i, t := range sTags {
						names[i] = t.Name
					}
					tagStr = " [" + strings.Join(names, ",") + "]"
				}
				fmt.Fprintf(w, "%s%s\t%s\t%s\t%d\t%s\n",
					truncate(name, 30), tagStr,
					truncate(s.ProjectDir, 20),
					truncate(branch, 15),
					s.MessageCount,
					formatRelativeTime(s.UpdatedAt),
				)
			}
			return w.Flush()
		},
	}

	cmd.Flags().BoolVarP(&all, "all", "a", false, "list sessions from all projects")
	cmd.Flags().StringVarP(&project, "project", "p", "", "filter by project directory")
	cmd.Flags().StringSliceVarP(&tags, "tag", "t", nil, "filter by tag (repeatable)")
	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "max results")
	cmd.Flags().StringVar(&sortBy, "sort", "updated", "sort by: updated, created, name")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "output as JSON")

	return cmd
}

// detectProjectDir returns the sanitized project directory name for the current
// working directory, matching how Claude Code names its project directories.
// Claude Code replaces all path separators with dashes, e.g.
// /mnt/c/Users/foo/project -> -mnt-c-Users-foo-project
func detectProjectDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(cwd, "/", "-"), nil
}
