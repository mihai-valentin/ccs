package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "ccs",
		Short: "Claude Code Session Manager",
		Long:  "A CLI tool for managing Claude Code sessions — list, search, tag, and resume sessions across projects.",
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
