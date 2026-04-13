package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/mihai/ccs/internal/format"
	"github.com/mihai/ccs/internal/model"
)

const timeFormat = time.RFC3339Nano

func formatTime(t time.Time) string {
	return t.Format(timeFormat)
}

func parseTime(s string) time.Time {
	return format.ParseTime(s)
}

func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func fromNull(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// upsertSessionSQL is the shared upsert query used by both UpsertSession and UpsertSessionTx.
const upsertSessionSQL = `
	INSERT INTO sessions (id, project_dir, cwd, git_branch, name, first_message, last_message, message_count, created_at, updated_at, file_size, file_mod_time)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		project_dir = excluded.project_dir,
		cwd = excluded.cwd,
		git_branch = excluded.git_branch,
		name = excluded.name,
		first_message = excluded.first_message,
		last_message = excluded.last_message,
		message_count = excluded.message_count,
		created_at = excluded.created_at,
		updated_at = excluded.updated_at,
		file_size = excluded.file_size,
		file_mod_time = excluded.file_mod_time`

func upsertSessionArgs(s *model.Session) []any {
	return []any{
		s.ID, s.ProjectDir, s.Cwd, nullStr(s.GitBranch), nullStr(s.Name),
		nullStr(s.FirstMessage), nullStr(s.LastMessage), s.MessageCount,
		formatTime(s.CreatedAt), formatTime(s.UpdatedAt), s.FileSize, formatTime(s.FileModTime),
	}
}

// UpsertSession inserts or updates a session record.
func (d *DB) UpsertSession(s *model.Session) error {
	_, err := d.Exec(upsertSessionSQL, upsertSessionArgs(s)...)
	return err
}

// UpsertSessionTx inserts or updates a session record within the given transaction.
func (d *DB) UpsertSessionTx(tx *sql.Tx, s *model.Session) error {
	_, err := tx.Exec(upsertSessionSQL, upsertSessionArgs(s)...)
	return err
}

// DeleteSession removes a session and its tag associations.
func (d *DB) DeleteSession(id string) error {
	_, err := d.Exec("DELETE FROM sessions WHERE id = ?", id)
	return err
}

// GetSessionByID retrieves a single session by its ID.
func (d *DB) GetSessionByID(id string) (*model.Session, error) {
	s := &model.Session{}
	var gitBranch, name, firstMsg, lastMsg, summary sql.NullString
	var createdAt, updatedAt, fileModTime string
	err := d.QueryRow(`
		SELECT id, project_dir, cwd, git_branch, name, first_message, last_message,
		       message_count, created_at, updated_at, file_size, file_mod_time, summary
		FROM sessions WHERE id = ?`, id).Scan(
		&s.ID, &s.ProjectDir, &s.Cwd, &gitBranch, &name,
		&firstMsg, &lastMsg, &s.MessageCount,
		&createdAt, &updatedAt, &s.FileSize, &fileModTime, &summary,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s.GitBranch = fromNull(gitBranch)
	s.Name = fromNull(name)
	s.FirstMessage = fromNull(firstMsg)
	s.LastMessage = fromNull(lastMsg)
	s.Summary = fromNull(summary)
	s.CreatedAt = format.ParseTime(createdAt)
	s.UpdatedAt = format.ParseTime(updatedAt)
	s.FileModTime = format.ParseTime(fileModTime)
	return s, nil
}

// ListSessions returns sessions matching the given filters.
func (d *DB) ListSessions(f model.SessionFilter) ([]model.Session, error) {
	query := `SELECT s.id, s.project_dir, s.cwd, s.git_branch, s.name, s.first_message,
	                 s.last_message, s.message_count, s.created_at, s.updated_at,
	                 s.file_size, s.file_mod_time, s.summary
	          FROM sessions s`
	var args []any
	var conditions []string

	if f.ProjectDir != "" {
		conditions = append(conditions, "s.project_dir = ?")
		args = append(args, f.ProjectDir)
	}

	if len(f.Tags) > 0 {
		placeholders := make([]string, len(f.Tags))
		for i, tag := range f.Tags {
			placeholders[i] = "?"
			args = append(args, tag)
		}
		query += fmt.Sprintf(` JOIN session_tags st ON s.id = st.session_id
		                        JOIN tags t ON st.tag_id = t.id AND t.name IN (%s)`,
			strings.Join(placeholders, ","))
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	switch f.SortBy {
	case "created":
		query += " ORDER BY s.created_at DESC"
	case "name":
		query += " ORDER BY s.name ASC"
	default:
		query += " ORDER BY s.updated_at DESC"
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 20
	}
	query += fmt.Sprintf(" LIMIT %d", limit)

	rows, err := d.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSessions(rows)
}

// SearchSessions performs a text search across name, messages, cwd, and branch.
func (d *DB) SearchSessions(query string) ([]model.Session, error) {
	pattern := "%" + query + "%"
	rows, err := d.Query(`
		SELECT id, project_dir, cwd, git_branch, name, first_message, last_message,
		       message_count, created_at, updated_at, file_size, file_mod_time, summary
		FROM sessions
		WHERE name LIKE ? OR first_message LIKE ? OR last_message LIKE ?
		      OR cwd LIKE ? OR git_branch LIKE ?
		ORDER BY updated_at DESC
		LIMIT 50`,
		pattern, pattern, pattern, pattern, pattern,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSessions(rows)
}

// AddTag creates a tag if it doesn't exist and associates it with the session.
func (d *DB) AddTag(sessionID, tagName string) error {
	_, err := d.Exec("INSERT OR IGNORE INTO tags (name) VALUES (?)", tagName)
	if err != nil {
		return err
	}
	_, err = d.Exec(`
		INSERT OR IGNORE INTO session_tags (session_id, tag_id)
		SELECT ?, id FROM tags WHERE name = ?`, sessionID, tagName)
	return err
}

// RemoveTag removes the association between a session and a tag.
func (d *DB) RemoveTag(sessionID, tagName string) error {
	_, err := d.Exec(`
		DELETE FROM session_tags WHERE session_id = ?
		AND tag_id = (SELECT id FROM tags WHERE name = ?)`, sessionID, tagName)
	return err
}

// ListTags returns all tags with their session counts.
func (d *DB) ListTags() ([]model.Tag, error) {
	rows, err := d.Query(`
		SELECT t.id, t.name FROM tags t
		ORDER BY t.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

// GetSessionTags returns all tags for a given session.
func (d *DB) GetSessionTags(sessionID string) ([]model.Tag, error) {
	rows, err := d.Query(`
		SELECT t.id, t.name FROM tags t
		JOIN session_tags st ON t.id = st.tag_id
		WHERE st.session_id = ?
		ORDER BY t.name`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

// ListProjects returns distinct project directories from indexed sessions.
func (d *DB) ListProjects() ([]string, error) {
	rows, err := d.Query("SELECT DISTINCT project_dir FROM sessions ORDER BY project_dir")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

// SessionMeta holds the minimal file-level metadata needed for incremental
// indexing change detection.
type SessionMeta struct {
	FileSize    int64
	FileModTime time.Time
}

// GetAllSessionMeta returns file-size and mod-time for every indexed session,
// keyed by session ID. Unlike ListSessions it applies no LIMIT, so the caller
// gets a complete picture regardless of how many sessions exist.
func (d *DB) GetAllSessionMeta() (map[string]SessionMeta, error) {
	rows, err := d.Query("SELECT id, file_size, file_mod_time FROM sessions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[string]SessionMeta)
	for rows.Next() {
		var id string
		var fileSize int64
		var fileModTime string
		if err := rows.Scan(&id, &fileSize, &fileModTime); err != nil {
			return nil, err
		}
		m[id] = SessionMeta{
			FileSize:    fileSize,
			FileModTime: format.ParseTime(fileModTime),
		}
	}
	return m, rows.Err()
}

// PurgeMissingSessions removes sessions whose IDs are not in the provided list.
// If existingIDs is empty, we return early without deleting anything. An empty
// list most likely means the scan found no files (e.g. ~/.claude was temporarily
// inaccessible), and blindly deleting all sessions would cause total data loss
// including cascade-deletion of all tags.
func (d *DB) PurgeMissingSessions(existingIDs []string) error {
	if len(existingIDs) == 0 {
		return nil
	}

	placeholders := make([]string, len(existingIDs))
	args := make([]any, len(existingIDs))
	for i, id := range existingIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	_, err := d.Exec(
		fmt.Sprintf("DELETE FROM sessions WHERE id NOT IN (%s)", strings.Join(placeholders, ",")),
		args...,
	)
	return err
}

// PurgeMissingSessionsTx is like PurgeMissingSessions but runs within the given transaction.
func (d *DB) PurgeMissingSessionsTx(tx *sql.Tx, existingIDs []string) error {
	if len(existingIDs) == 0 {
		return nil
	}

	placeholders := make([]string, len(existingIDs))
	args := make([]any, len(existingIDs))
	for i, id := range existingIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	_, err := tx.Exec(
		fmt.Sprintf("DELETE FROM sessions WHERE id NOT IN (%s)", strings.Join(placeholders, ",")),
		args...,
	)
	return err
}

// UpdateSummary stores an AI-generated summary for the given session.
func (d *DB) UpdateSummary(sessionID, summary string) error {
	_, err := d.Exec("UPDATE sessions SET summary = ? WHERE id = ?", nullStr(summary), sessionID)
	return err
}

func scanSessions(rows *sql.Rows) ([]model.Session, error) {
	var sessions []model.Session
	for rows.Next() {
		var s model.Session
		var gitBranch, name, firstMsg, lastMsg, summary sql.NullString
		var createdAt, updatedAt, fileModTime string
		if err := rows.Scan(
			&s.ID, &s.ProjectDir, &s.Cwd, &gitBranch, &name,
			&firstMsg, &lastMsg, &s.MessageCount,
			&createdAt, &updatedAt, &s.FileSize, &fileModTime, &summary,
		); err != nil {
			return nil, err
		}
		s.GitBranch = fromNull(gitBranch)
		s.Name = fromNull(name)
		s.FirstMessage = fromNull(firstMsg)
		s.LastMessage = fromNull(lastMsg)
		s.Summary = fromNull(summary)
		s.CreatedAt = format.ParseTime(createdAt)
		s.UpdatedAt = format.ParseTime(updatedAt)
		s.FileModTime = format.ParseTime(fileModTime)
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}
