package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/mihai/ccs/internal/db"
	"github.com/mihai/ccs/internal/indexer"
)

var (
	flagDBPath    string
	flagClaudeDir string
)

func defaultDBPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ccs", "ccs.db")
}

func defaultClaudeDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude")
}

func getDBPath() string {
	if flagDBPath != "" {
		return flagDBPath
	}
	return defaultDBPath()
}

func getClaudeDir() string {
	if flagClaudeDir != "" {
		return flagClaudeDir
	}
	return defaultClaudeDir()
}

func openDB() (*db.DB, error) {
	return db.Open(getDBPath())
}

func syncIndex() (*db.DB, error) {
	d, err := openDB()
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	idx := indexer.NewIndexer(d, getClaudeDir())
	if err := idx.Run(); err != nil {
		d.Close()
		return nil, fmt.Errorf("syncing index: %w", err)
	}
	return d, nil
}

func formatRelativeTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		return fmt.Sprintf("%dm ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		return fmt.Sprintf("%dh ago", h)
	case d < 30*24*time.Hour:
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	case d < 365*24*time.Hour:
		months := int(d.Hours() / 24 / 30)
		return fmt.Sprintf("%dmo ago", months)
	default:
		years := int(d.Hours() / 24 / 365)
		return fmt.Sprintf("%dy ago", years)
	}
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
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func sessionDisplayName(name, id string) string {
	if name != "" {
		return name
	}
	if len(id) >= 8 {
		return id[:8]
	}
	return id
}
