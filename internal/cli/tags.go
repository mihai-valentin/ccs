package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

type tagWithCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func newTagsCmd() *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "tags",
		Short: "List all tags with session counts",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := openDB()
			if err != nil {
				return err
			}
			defer d.Close()

			rows, err := d.Query(`
				SELECT t.name, COUNT(st.session_id) as cnt
				FROM tags t
				LEFT JOIN session_tags st ON t.id = st.tag_id
				GROUP BY t.id, t.name
				ORDER BY t.name`)
			if err != nil {
				return fmt.Errorf("listing tags: %w", err)
			}
			defer rows.Close()

			var tags []tagWithCount
			for rows.Next() {
				var t tagWithCount
				if err := rows.Scan(&t.Name, &t.Count); err != nil {
					return err
				}
				tags = append(tags, t)
			}
			if err := rows.Err(); err != nil {
				return err
			}

			if jsonOut {
				return printJSON(tags)
			}

			if len(tags) == 0 {
				fmt.Println("No tags found.")
				return nil
			}

			w := newTabWriter()
			fmt.Fprintln(w, "TAG\tSESSIONS")
			for _, t := range tags {
				fmt.Fprintf(w, "%s\t%d\n", t.Name, t.Count)
			}
			return w.Flush()
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "output as JSON")
	return cmd
}
