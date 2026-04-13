package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// DB wraps an *sql.DB connection to the ccs SQLite database.
type DB struct {
	*sql.DB
}

// Open opens (or creates) the SQLite database at dbPath, runs schema
// migrations, and returns a *DB handle. Parent directories are created
// if they don't exist.
func Open(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Force a single connection so that connection-scoped PRAGMAs (like
	// foreign_keys=ON) apply to every query. sql.Open returns a pool and
	// PRAGMAs set on one connection don't carry over to new ones. A single
	// connection is fine for a local CLI tool with no concurrent DB access.
	sqlDB.SetMaxOpenConns(1)

	// Enable WAL mode and foreign keys.
	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
	} {
		if _, err := sqlDB.Exec(pragma); err != nil {
			sqlDB.Close()
			return nil, fmt.Errorf("exec %s: %w", pragma, err)
		}
	}

	if _, err := sqlDB.Exec(schemaSQL); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("run schema migration: %w", err)
	}

	// Migrate: add summary column if missing (for existing DBs).
	if _, err := sqlDB.Exec("ALTER TABLE sessions ADD COLUMN summary TEXT"); err != nil {
		if !strings.Contains(err.Error(), "duplicate column name") {
			sqlDB.Close()
			return nil, fmt.Errorf("alter table add summary column: %w", err)
		}
	}

	return &DB{sqlDB}, nil
}
