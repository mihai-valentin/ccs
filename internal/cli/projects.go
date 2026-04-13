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

			dbProjects, err := d.ListProjectsWithCounts()
			if err != nil {
				return fmt.Errorf("listing projects: %w", err)
			}

			projects := make([]projectWithCount, len(dbProjects))
			for i, p := range dbProjects {
				projects[i] = projectWithCount{Name: p.Name, Count: p.Count}
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
