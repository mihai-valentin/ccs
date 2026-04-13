package db

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/mihai-valentin/ccs/internal/model"
)

func setupTestDB(t *testing.T) *DB {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	d, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func insertTestSession(t *testing.T, d *DB, id string) {
	t.Helper()
	now := time.Now()
	s := &model.Session{
		ID:          id,
		ProjectDir:  "test-project",
		Cwd:         "/tmp/test",
		MessageCount: 1,
		CreatedAt:   now,
		UpdatedAt:   now,
		FileSize:    100,
		FileModTime: now,
	}
	if err := d.UpsertSession(s); err != nil {
		t.Fatalf("insert test session %s: %v", id, err)
	}
}

func TestFileModTimeNanosecondPrecision(t *testing.T) {
	d := setupTestDB(t)

	// A time with nanosecond precision (not on a second boundary).
	modTime := time.Date(2026, 4, 13, 10, 30, 45, 123456789, time.UTC)

	s := &model.Session{
		ID:           "nano-session",
		ProjectDir:   "test-project",
		Cwd:          "/tmp/test",
		MessageCount: 1,
		CreatedAt:    modTime,
		UpdatedAt:    modTime,
		FileSize:     200,
		FileModTime:  modTime,
	}
	if err := d.UpsertSession(s); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, err := d.GetSessionByID("nano-session")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil {
		t.Fatal("session not found")
	}
	if !got.FileModTime.Equal(modTime) {
		t.Errorf("FileModTime lost precision: stored %v, got %v", modTime, got.FileModTime)
	}
}

func TestParseTimeFallsBackToRFC3339(t *testing.T) {
	// Simulate a value stored before the RFC3339Nano fix.
	oldFormat := "2026-04-13T10:30:45Z"
	got := parseTime(oldFormat)
	want := time.Date(2026, 4, 13, 10, 30, 45, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("parseTime fallback failed: got %v, want %v", got, want)
	}
}

func TestPurgeMissingSessions_EmptySliceDoesNotDelete(t *testing.T) {
	d := setupTestDB(t)

	// Insert some sessions.
	insertTestSession(t, d, "session-1")
	insertTestSession(t, d, "session-2")
	insertTestSession(t, d, "session-3")

	// Add a tag to verify cascade deletion doesn't happen.
	if err := d.AddTag("session-1", "important"); err != nil {
		t.Fatalf("add tag: %v", err)
	}

	// Purge with empty slice — should be a no-op.
	if err := d.PurgeMissingSessions(nil); err != nil {
		t.Fatalf("purge: %v", err)
	}

	// All sessions must still exist.
	for _, id := range []string{"session-1", "session-2", "session-3"} {
		s, err := d.GetSessionByID(id)
		if err != nil {
			t.Fatalf("get session %s: %v", id, err)
		}
		if s == nil {
			t.Errorf("session %s was deleted by PurgeMissingSessions(nil)", id)
		}
	}

	// Tag must still exist.
	tags, err := d.GetSessionTags("session-1")
	if err != nil {
		t.Fatalf("get tags: %v", err)
	}
	if len(tags) != 1 || tags[0].Name != "important" {
		t.Errorf("tag was lost after PurgeMissingSessions(nil), got %v", tags)
	}
}

func TestPurgeMissingSessions_EmptyExplicitSlice(t *testing.T) {
	d := setupTestDB(t)

	insertTestSession(t, d, "session-a")

	// Explicitly empty (not nil) slice — same behavior.
	if err := d.PurgeMissingSessions([]string{}); err != nil {
		t.Fatalf("purge: %v", err)
	}

	s, err := d.GetSessionByID("session-a")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if s == nil {
		t.Error("session was deleted by PurgeMissingSessions([]string{})")
	}
}

func TestPurgeMissingSessions_DeletesAbsentSessions(t *testing.T) {
	d := setupTestDB(t)

	insertTestSession(t, d, "keep-1")
	insertTestSession(t, d, "keep-2")
	insertTestSession(t, d, "remove-1")

	if err := d.PurgeMissingSessions([]string{"keep-1", "keep-2"}); err != nil {
		t.Fatalf("purge: %v", err)
	}

	s, err := d.GetSessionByID("remove-1")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if s != nil {
		t.Error("expected remove-1 to be purged")
	}

	for _, id := range []string{"keep-1", "keep-2"} {
		s, err := d.GetSessionByID(id)
		if err != nil {
			t.Fatalf("get session %s: %v", id, err)
		}
		if s == nil {
			t.Errorf("session %s should not have been purged", id)
		}
	}
}
