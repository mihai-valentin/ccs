package cli

import (
	"fmt"

	"github.com/mihai/ccs/internal/indexer"
	"github.com/spf13/cobra"
)

func newReindexCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reindex",
		Short: "Force a full re-index of all sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := openDB()
			if err != nil {
				return err
			}
			defer d.Close()

			claudeDir, err := getClaudeDir()
			if err != nil {
				return err
			}
			idx := indexer.NewIndexer(d, claudeDir)
			if err := idx.Reindex(); err != nil {
				return fmt.Errorf("reindexing: %w", err)
			}

			fmt.Println("Reindex complete.")
			return nil
		},
	}
	return cmd
}
