package opener

import (
	"fmt"
	"strings"

	"github.com/mihai/ccs/internal/db"
	"github.com/mihai/ccs/internal/format"
	"github.com/mihai/ccs/internal/model"
)

// ErrAmbiguous is returned when a session identifier matches multiple sessions.
type ErrAmbiguous struct {
	Identifier string
	Matches    []model.Session
}

func (e *ErrAmbiguous) Error() string {
	candidates := make([]string, len(e.Matches))
	for i, m := range e.Matches {
		label := m.ID
		if m.Name != "" {
			label = fmt.Sprintf("%s (%s)", m.ID, m.Name)
		}
		candidates[i] = label
	}
	return fmt.Sprintf("identifier %q is ambiguous — matches %d sessions: %s",
		e.Identifier, len(e.Matches), strings.Join(candidates, ", "))
}

// ErrNotFound is returned when no session matches the given identifier.
type ErrNotFound struct {
	Identifier string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("no session found matching %q", e.Identifier)
}

// ResolveSession resolves a user-supplied identifier to a single session.
// Resolution order: exact UUID > UUID prefix (4+ chars) > exact name > fuzzy name match.
func ResolveSession(d *db.DB, identifier string) (*model.Session, error) {
	// 1. Exact UUID match.
	s, err := d.GetSessionByID(identifier)
	if err != nil {
		return nil, fmt.Errorf("lookup by ID: %w", err)
	}
	if s != nil {
		return s, nil
	}

	// 2. UUID prefix match (require at least 4 characters to avoid noise).
	if len(identifier) >= 4 {
		matches, err := queryByIDPrefix(d, identifier)
		if err != nil {
			return nil, fmt.Errorf("lookup by ID prefix: %w", err)
		}
		if len(matches) == 1 {
			return &matches[0], nil
		}
		if len(matches) > 1 {
			return nil, &ErrAmbiguous{Identifier: identifier, Matches: matches}
		}
	}

	// 3. Exact name match.
	matches, err := queryByExactName(d, identifier)
	if err != nil {
		return nil, fmt.Errorf("lookup by name: %w", err)
	}
	if len(matches) == 1 {
		return &matches[0], nil
	}
	if len(matches) > 1 {
		return nil, &ErrAmbiguous{Identifier: identifier, Matches: matches}
	}

	// 4. Fuzzy name match (case-insensitive substring).
	matches, err = queryByFuzzyName(d, identifier)
	if err != nil {
		return nil, fmt.Errorf("fuzzy lookup: %w", err)
	}
	if len(matches) == 1 {
		return &matches[0], nil
	}
	if len(matches) > 1 {
		return nil, &ErrAmbiguous{Identifier: identifier, Matches: matches}
	}

	return nil, &ErrNotFound{Identifier: identifier}
}

func queryByIDPrefix(d *db.DB, prefix string) ([]model.Session, error) {
	// Use LIKE with the prefix escaped to avoid SQL injection via % or _ chars.
	escaped := strings.ReplaceAll(strings.ReplaceAll(prefix, "%", "\\%"), "_", "\\_")
	rows, err := d.Query(`
		SELECT id, project_dir, cwd, git_branch, name, first_message, last_message,
		       message_count, created_at, updated_at, file_size, file_mod_time, summary
		FROM sessions
		WHERE id LIKE ? ESCAPE '\'
		ORDER BY updated_at DESC
		LIMIT 10`, escaped+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanResolveRows(rows)
}

func queryByExactName(d *db.DB, name string) ([]model.Session, error) {
	rows, err := d.Query(`
		SELECT id, project_dir, cwd, git_branch, name, first_message, last_message,
		       message_count, created_at, updated_at, file_size, file_mod_time, summary
		FROM sessions
		WHERE name = ?
		ORDER BY updated_at DESC
		LIMIT 10`, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanResolveRows(rows)
}

func queryByFuzzyName(d *db.DB, term string) ([]model.Session, error) {
	pattern := "%" + strings.ReplaceAll(strings.ReplaceAll(term, "%", "\\%"), "_", "\\_") + "%"
	rows, err := d.Query(`
		SELECT id, project_dir, cwd, git_branch, name, first_message, last_message,
		       message_count, created_at, updated_at, file_size, file_mod_time, summary
		FROM sessions
		WHERE name LIKE ? ESCAPE '\'
		ORDER BY updated_at DESC
		LIMIT 10`, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanResolveRows(rows)
}

func scanResolveRows(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]model.Session, error) {
	var sessions []model.Session
	for rows.Next() {
		var s model.Session
		var gitBranch, name, firstMsg, lastMsg, summary nullString
		var createdAt, updatedAt, fileModTime string
		if err := rows.Scan(
			&s.ID, &s.ProjectDir, &s.Cwd, &gitBranch, &name,
			&firstMsg, &lastMsg, &s.MessageCount,
			&createdAt, &updatedAt, &s.FileSize, &fileModTime, &summary,
		); err != nil {
			return nil, err
		}
		s.GitBranch = gitBranch.String()
		s.Name = name.String()
		s.FirstMessage = firstMsg.String()
		s.LastMessage = lastMsg.String()
		s.Summary = summary.String()
		s.CreatedAt = format.ParseTime(createdAt)
		s.UpdatedAt = format.ParseTime(updatedAt)
		s.FileModTime = format.ParseTime(fileModTime)
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// nullString is a minimal nullable string scanner for SQL results.
type nullString struct {
	val   string
	valid bool
}

func (ns *nullString) Scan(src any) error {
	if src == nil {
		ns.val = ""
		ns.valid = false
		return nil
	}
	switch v := src.(type) {
	case string:
		ns.val = v
		ns.valid = true
	case []byte:
		ns.val = string(v)
		ns.valid = true
	default:
		return fmt.Errorf("nullString.Scan: unsupported type %T", src)
	}
	return nil
}

func (ns nullString) String() string {
	return ns.val
}

