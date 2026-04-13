package cli

import (
	"github.com/mihai-valentin/ccs/internal/opener"
	"github.com/spf13/cobra"
)

func newOpenCmd() *cobra.Command {
	var newTerminal bool

	cmd := &cobra.Command{
		Use:   "open <id|name>",
		Short: "Open a session (cd into project dir + resume)",
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

			return opener.OpenSession(*session, newTerminal)
		},
	}

	cmd.Flags().BoolVar(&newTerminal, "new-terminal", false, "spawn in a new terminal")

	return cmd
}
