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

			dbTags, err := d.ListTagsWithCounts()
			if err != nil {
				return fmt.Errorf("listing tags: %w", err)
			}

			tags := make([]tagWithCount, len(dbTags))
			for i, t := range dbTags {
				tags[i] = tagWithCount{Name: t.Name, Count: t.Count}
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
