package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/mihai/ccs/internal/db"
	"github.com/mihai/ccs/internal/format"
	"github.com/mihai/ccs/internal/indexer"
)

var (
	flagDBPath    string
	flagClaudeDir string
)

func defaultDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "ccs", "ccs.db"), nil
}

func defaultClaudeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".claude"), nil
}

func getDBPath() (string, error) {
	if flagDBPath != "" {
		return flagDBPath, nil
	}
	return defaultDBPath()
}

func getClaudeDir() (string, error) {
	if flagClaudeDir != "" {
		return flagClaudeDir, nil
	}
	return defaultClaudeDir()
}

func openDB() (*db.DB, error) {
	dbPath, err := getDBPath()
	if err != nil {
		return nil, err
	}
	return db.Open(dbPath)
}

func syncIndex() (*db.DB, error) {
	d, err := openDB()
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	claudeDir, err := getClaudeDir()
	if err != nil {
		d.Close()
		return nil, err
	}
	idx := indexer.NewIndexer(d, claudeDir)
	if err := idx.Run(); err != nil {
		d.Close()
		return nil, fmt.Errorf("syncing index: %w", err)
	}
	return d, nil
}

func formatRelativeTime(t time.Time) string {
	return format.FormatRelativeTime(t)
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func newTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
}

func truncate(s string, maxLen int) string {
	return format.Truncate(s, maxLen)
}

func sessionDisplayName(name, id string) string {
	return format.SessionDisplayName(name, id)
}
