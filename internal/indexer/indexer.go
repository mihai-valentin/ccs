package indexer

import (
	"fmt"
	"log"

	"github.com/mihai-valentin/ccs/internal/db"
	"github.com/mihai-valentin/ccs/internal/model"
)

// Indexer scans for Claude Code session files and syncs metadata into the database.
type Indexer struct {
	db        *db.DB
	claudeDir string
}

// NewIndexer creates an Indexer that scans claudeDir for session files
// and stores metadata in the given database.
func NewIndexer(database *db.DB, claudeDir string) *Indexer {
	return &Indexer{
		db:        database,
		claudeDir: claudeDir,
	}
}

// Run performs an incremental index: scan files, diff against DB records,
// re-parse only changed/new files, upsert into DB, and purge deleted sessions.
func (idx *Indexer) Run() error {
	return idx.run(false)
}

// Reindex forces a full re-parse of all session files regardless of
// file_size/file_mod_time changes.
func (idx *Indexer) Reindex() error {
	return idx.run(true)
}

func (idx *Indexer) run(force bool) error {
	files, scanHadErrors, err := ScanSessions(idx.claudeDir)
	if err != nil {
		return fmt.Errorf("scan sessions: %w", err)
	}

	// Build a set of discovered session IDs for purging later.
	existingIDs := make([]string, 0, len(files))
	for _, f := range files {
		existingIDs = append(existingIDs, f.SessionID)
	}

	// Load existing DB records to detect changes.
	var dbMeta map[string]db.SessionMeta
	if !force {
		dbMeta, err = idx.db.GetAllSessionMeta()
		if err != nil {
			return fmt.Errorf("load existing sessions: %w", err)
		}
	}

	// Wrap all upserts and the purge in a single transaction so the DB
	// is never left in a half-updated state (e.g. if the process is killed
	// mid-index or a write fails).
	tx, err := idx.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() // no-op after Commit

	for _, f := range files {
		if !force {
			if existing, ok := dbMeta[f.SessionID]; ok {
				if existing.FileSize == f.FileSize && existing.FileModTime.Equal(f.FileModTime) {
					continue // unchanged, skip
				}
			}
		}

		parsed, err := ParseSessionFile(f.Path)
		if err != nil {
			log.Printf("warning: failed to parse %s: %v", f.Path, err)
			continue
		}
		if parsed == nil {
			continue // empty/unparseable file
		}

		session := &model.Session{
			ID:           parsed.SessionID,
			ProjectDir:   f.ProjectDir,
			Cwd:          parsed.Cwd,
			GitBranch:    parsed.GitBranch,
			Name:         parsed.Name,
			FirstMessage: parsed.FirstMessage,
			LastMessage:  parsed.LastMessage,
			MessageCount: parsed.MessageCount,
			CreatedAt:    parsed.FirstTime,
			UpdatedAt:    parsed.LastTime,
			FileSize:     f.FileSize,
			FileModTime:  f.FileModTime,
		}

		if err := idx.db.UpsertSessionTx(tx, session); err != nil {
			log.Printf("warning: failed to upsert session %s: %v", parsed.SessionID, err)
			continue
		}
	}

	// Purge sessions whose files no longer exist on disk.
	// Skip purge if the scan encountered errors (partial results) — purging
	// based on an incomplete scan would incorrectly delete sessions whose
	// files were simply inaccessible.
	if scanHadErrors {
		log.Printf("warning: skipping purge of missing sessions because scan encountered errors (partial results)")
	} else {
		if err := idx.db.PurgeMissingSessionsTx(tx, existingIDs); err != nil {
			return fmt.Errorf("purge missing sessions: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
