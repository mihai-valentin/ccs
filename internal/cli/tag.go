package cli

import (
	"fmt"
	"strings"

	"github.com/mihai/ccs/internal/opener"
	"github.com/spf13/cobra"
)

func newTagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag <id|name> <tag> [tags...]",
		Short: "Add tags to a session",
		Args:  cobra.MinimumNArgs(2),
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

			for _, tag := range args[1:] {
				if err := d.AddTag(session.ID, tag); err != nil {
					return fmt.Errorf("adding tag %q: %w", tag, err)
				}
			}

			name := sessionDisplayName(session.Name, session.ID)
			fmt.Printf("Tagged %s with: %s\n", name, strings.Join(args[1:], ", "))
			return nil
		},
	}
	return cmd
}

func newUntagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "untag <id|name> <tag> [tags...]",
		Short: "Remove tags from a session",
		Args:  cobra.MinimumNArgs(2),
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

			for _, tag := range args[1:] {
				if err := d.RemoveTag(session.ID, tag); err != nil {
					return fmt.Errorf("removing tag %q: %w", tag, err)
				}
			}

			name := sessionDisplayName(session.Name, session.ID)
			fmt.Printf("Removed tags from %s: %s\n", name, strings.Join(args[1:], ", "))
			return nil
		},
	}
	return cmd
}
