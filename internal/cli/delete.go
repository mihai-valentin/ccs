package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mihai/ccs/internal/opener"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <id|name>",
		Short: "Delete a session (JSONL file + index entry)",
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

			name := sessionDisplayName(session.Name, session.ID)

			if !force {
				fmt.Printf("Delete session %q (%s)? [y/N] ", name, session.ID)
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))
				if answer != "y" && answer != "yes" {
					fmt.Println("Aborted.")
					return nil
				}
			}

			// Delete the JSONL file from disk
			claudeDir, err := getClaudeDir()
			if err != nil {
				return err
			}
			jsonlPath := filepath.Join(claudeDir, "projects", session.ProjectDir, session.ID+".jsonl")
			if err := os.Remove(jsonlPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("removing JSONL file: %w", err)
			}

			// Remove from database index
			if err := d.DeleteSession(session.ID); err != nil {
				return fmt.Errorf("removing from index: %w", err)
			}

			fmt.Printf("Deleted session %s\n", name)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation prompt")
	return cmd
}
