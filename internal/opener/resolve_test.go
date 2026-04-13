package opener

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mihai-valentin/ccs/internal/db"
	"github.com/mihai-valentin/ccs/internal/model"
)

func setupTestDB(t *testing.T) *db.DB {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	d, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func seedSession(t *testing.T, d *db.DB, id, name, cwd string) {
	t.Helper()
	now := time.Now()
	s := &model.Session{
		ID:          id,
		ProjectDir:  "/projects/test",
		Cwd:         cwd,
		Name:        name,
		MessageCount: 1,
		CreatedAt:   now,
		UpdatedAt:   now,
		FileSize:    100,
		FileModTime: now,
	}
	if err := d.UpsertSession(s); err != nil {
		t.Fatalf("seed session %s: %v", id, err)
	}
}

func TestResolveSession_ExactID(t *testing.T) {
	d := setupTestDB(t)
	seedSession(t, d, "aaaa-bbbb-cccc-dddd", "my-session", "/tmp")

	s, err := ResolveSession(d, "aaaa-bbbb-cccc-dddd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.ID != "aaaa-bbbb-cccc-dddd" {
		t.Errorf("got ID %q, want %q", s.ID, "aaaa-bbbb-cccc-dddd")
	}
}

func TestResolveSession_IDPrefix(t *testing.T) {
	d := setupTestDB(t)
	seedSession(t, d, "abcd-1234-5678-9999", "session-one", "/tmp")
	seedSession(t, d, "ffff-0000-1111-2222", "session-two", "/tmp")

	s, err := ResolveSession(d, "abcd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.ID != "abcd-1234-5678-9999" {
		t.Errorf("got ID %q, want %q", s.ID, "abcd-1234-5678-9999")
	}
}

func TestResolveSession_IDPrefixAmbiguous(t *testing.T) {
	d := setupTestDB(t)
	seedSession(t, d, "abcd-1111-0000-0000", "s1", "/tmp")
	seedSession(t, d, "abcd-2222-0000-0000", "s2", "/tmp")

	_, err := ResolveSession(d, "abcd")
	if err == nil {
		t.Fatal("expected error for ambiguous prefix, got nil")
	}
	ambErr, ok := err.(*ErrAmbiguous)
	if !ok {
		t.Fatalf("expected *ErrAmbiguous, got %T: %v", err, err)
	}
	if len(ambErr.Matches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(ambErr.Matches))
	}
	// Error message must list the candidate session IDs.
	msg := ambErr.Error()
	if !strings.Contains(msg, "abcd-1111") || !strings.Contains(msg, "abcd-2222") {
		t.Errorf("error message %q should list candidate IDs", msg)
	}
}

func TestResolveSession_ExactName(t *testing.T) {
	d := setupTestDB(t)
	seedSession(t, d, "1111-2222-3333-4444", "deploy-fix", "/tmp")

	s, err := ResolveSession(d, "deploy-fix")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.ID != "1111-2222-3333-4444" {
		t.Errorf("got ID %q, want %q", s.ID, "1111-2222-3333-4444")
	}
}

func TestResolveSession_FuzzyName(t *testing.T) {
	d := setupTestDB(t)
	seedSession(t, d, "5555-6666-7777-8888", "refactor-auth-module", "/tmp")

	s, err := ResolveSession(d, "auth")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.ID != "5555-6666-7777-8888" {
		t.Errorf("got ID %q, want %q", s.ID, "5555-6666-7777-8888")
	}
}

func TestResolveSession_NotFound(t *testing.T) {
	d := setupTestDB(t)

	_, err := ResolveSession(d, "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*ErrNotFound); !ok {
		t.Errorf("expected *ErrNotFound, got %T: %v", err, err)
	}
}

func TestResolveSession_ShortPrefixSkipped(t *testing.T) {
	d := setupTestDB(t)
	seedSession(t, d, "abc0-0000-0000-0000", "short", "/tmp")

	// 3-char prefix should skip prefix matching and fall through to name/fuzzy.
	_, err := ResolveSession(d, "abc")
	if err == nil {
		t.Fatal("expected error (short prefix should not match), got nil")
	}
	if _, ok := err.(*ErrNotFound); !ok {
		t.Errorf("expected *ErrNotFound, got %T: %v", err, err)
	}
}

func TestOpenSession_MissingCwd(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "ccs-test-nonexistent-dir-"+t.Name())
	// Ensure the directory does not exist.
	os.RemoveAll(dir)

	s := model.Session{
		ID:  "test-id",
		Cwd: dir,
	}

	err := OpenSession(s, false)
	if err == nil {
		t.Fatal("expected error for missing cwd, got nil")
	}
	if got := err.Error(); !strings.Contains(got, "no longer exists") {
		t.Errorf("error message %q should mention directory no longer exists", got)
	}
}
