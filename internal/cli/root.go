package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "ccs",
		Short: "Claude Code Session Manager",
		Long:  "A CLI tool for managing Claude Code sessions — list, search, tag, and resume sessions across projects.",
	}

	rootCmd.PersistentFlags().StringVar(&flagDBPath, "db-path", "", "path to SQLite database (default: ~/.config/ccs/ccs.db)")
	rootCmd.PersistentFlags().StringVar(&flagClaudeDir, "claude-dir", "", "path to Claude Code data directory (default: ~/.claude)")

	rootCmd.AddCommand(
		newListCmd(),
		newSearchCmd(),
		newShowCmd(),
		newOpenCmd(),
		newTagCmd(),
		newUntagCmd(),
		newTagsCmd(),
		newProjectsCmd(),
		newDeleteCmd(),
		newReindexCmd(),
		newCompletionCmd(),
		newUICmd(),
	)

	return rootCmd
}

func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
