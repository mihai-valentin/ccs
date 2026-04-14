package cli

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// ASCII squirrel by Erik Andersson (from ascii.co.uk/art/squirrel).
//
//go:embed mascot.txt
var mascotASCII string

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "ccs",
		Short: "Claude Code Session Manager",
		Long:  "A CLI tool for managing Claude Code sessions — list, search, tag, and resume sessions across projects.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), strings.TrimRight(mascotASCII, "\n"))
			fmt.Fprintln(cmd.OutOrStdout())
			_ = cmd.Help()
		},
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
		newSummarizeCmd(),
		newThemeCmd(),
	)

	return rootCmd
}

func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
