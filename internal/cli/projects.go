package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

type projectWithCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func newProjectsCmd() *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "projects",
		Short: "List all known projects with session counts",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := syncIndex()
			if err != nil {
				return err
			}
			defer d.Close()

			rows, err := d.Query(`
				SELECT project_dir, COUNT(*) as cnt
				FROM sessions
				GROUP BY project_dir
				ORDER BY project_dir`)
			if err != nil {
				return fmt.Errorf("listing projects: %w", err)
			}
			defer rows.Close()

			var projects []projectWithCount
			for rows.Next() {
				var p projectWithCount
				if err := rows.Scan(&p.Name, &p.Count); err != nil {
					return err
				}
				projects = append(projects, p)
			}
			if err := rows.Err(); err != nil {
				return err
			}

			if jsonOut {
				return printJSON(projects)
			}

			if len(projects) == 0 {
				fmt.Println("No projects found.")
				return nil
			}

			w := newTabWriter()
			fmt.Fprintln(w, "PROJECT\tSESSIONS")
			for _, p := range projects {
				fmt.Fprintf(w, "%s\t%d\n", p.Name, p.Count)
			}
			return w.Flush()
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "output as JSON")
	return cmd
}
