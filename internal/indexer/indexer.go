package indexer

import (
	"fmt"
	"log"

	"github.com/mihai/ccs/internal/db"
	"github.com/mihai/ccs/internal/model"
)

// Indexer scans for Claude Code session files and syncs metadata into the database.
type Indexer struct {
	db       *db.DB
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
	files, err := ScanSessions(idx.claudeDir)
	if err != nil {
		return fmt.Errorf("scan sessions: %w", err)
	}

	// Build a set of discovered session IDs for purging later.
	existingIDs := make([]string, 0, len(files))
	for _, f := range files {
		existingIDs = append(existingIDs, f.SessionID)
	}

	// Load existing DB records to detect changes.
	var dbSessions map[string]*model.Session
	if !force {
		dbSessions, err = idx.loadSessionMap()
		if err != nil {
			return fmt.Errorf("load existing sessions: %w", err)
		}
	}

	for _, f := range files {
		if !force {
			if existing, ok := dbSessions[f.SessionID]; ok {
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

		if err := idx.db.UpsertSession(session); err != nil {
			log.Printf("warning: failed to upsert session %s: %v", parsed.SessionID, err)
			continue
		}
	}

	// Purge sessions whose files no longer exist on disk.
	if err := idx.db.PurgeMissingSessions(existingIDs); err != nil {
		return fmt.Errorf("purge missing sessions: %w", err)
	}

	return nil
}

// loadSessionMap fetches all sessions from the DB keyed by ID.
func (idx *Indexer) loadSessionMap() (map[string]*model.Session, error) {
	sessions, err := idx.db.ListSessions(model.SessionFilter{Limit: 100000})
	if err != nil {
		return nil, err
	}
	m := make(map[string]*model.Session, len(sessions))
	for i := range sessions {
		m[sessions[i].ID] = &sessions[i]
	}
	return m, nil
}
