package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mihai/ccs/internal/db"
	"github.com/mihai/ccs/internal/model"
	"github.com/mihai/ccs/internal/ollama"
	"github.com/mihai/ccs/internal/opener"
	"github.com/mihai/ccs/internal/summarizer"
	"github.com/spf13/cobra"
)

func newSummarizeCmd() *cobra.Command {
	var all bool
	var ollamaModel string
	var ollamaURL string
	var force bool

	cmd := &cobra.Command{
		Use:   "summarize [id|name]",
		Short: "Generate AI summary for a session using local Ollama",
		Long:  "Uses a local Ollama model to summarize session conversations. Pass a session id/name, or --all to summarize all unsummarized sessions.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := syncIndex()
			if err != nil {
				return err
			}
			defer d.Close()

			client := ollama.NewClient(ollamaURL, ollamaModel)
			if err := client.Ping(); err != nil {
				return err
			}

			claudeDir := getClaudeDir()

			if all {
				return summarizeAllSessions(d, client, claudeDir, force)
			}

			if len(args) == 0 {
				return fmt.Errorf("provide a session id/name, or use --all")
			}

			session, err := opener.ResolveSession(d, args[0])
			if err != nil {
				return err
			}

			if session.Summary != "" && !force {
				fmt.Printf("Session already has a summary (use --force to regenerate):\n  %s\n", session.Summary)
				return nil
			}

			return doSummarize(d, client, claudeDir, session.ID, session.ProjectDir, sessionDisplayName(session.Name, session.ID))
		},
	}

	cmd.Flags().BoolVarP(&all, "all", "a", false, "summarize all sessions that don't have a summary yet")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "regenerate summary even if one exists")
	cmd.Flags().StringVar(&ollamaModel, "model", ollama.DefaultModel, "Ollama model to use")
	cmd.Flags().StringVar(&ollamaURL, "ollama-url", ollama.DefaultURL, "Ollama API base URL")

	return cmd
}

func doSummarize(d *db.DB, client *ollama.Client, claudeDir, sessionID, projectDir, displayName string) error {
	jsonlPath := filepath.Join(claudeDir, "projects", projectDir, sessionID+".jsonl")

	if _, err := os.Stat(jsonlPath); err != nil {
		return fmt.Errorf("session file not found: %s", jsonlPath)
	}

	fmt.Printf("Summarizing %s... ", displayName)

	summary, err := summarizer.Summarize(client, jsonlPath)
	if err != nil {
		fmt.Println("failed")
		return err
	}

	if err := d.UpdateSummary(sessionID, summary); err != nil {
		fmt.Println("failed")
		return fmt.Errorf("saving summary: %w", err)
	}

	fmt.Println("done")
	fmt.Printf("  %s\n", summary)
	return nil
}

func summarizeAllSessions(d *db.DB, client *ollama.Client, claudeDir string, force bool) error {
	sessions, err := d.ListSessions(model.SessionFilter{Limit: 10000})
	if err != nil {
		return err
	}

	var count int
	for _, s := range sessions {
		if s.Summary != "" && !force {
			continue
		}
		if err := doSummarize(d, client, claudeDir, s.ID, s.ProjectDir, sessionDisplayName(s.Name, s.ID)); err != nil {
			fmt.Fprintf(os.Stderr, "  warning: %v\n", err)
			continue
		}
		count++
	}

	fmt.Printf("\nSummarized %d session(s).\n", count)
	return nil
}
