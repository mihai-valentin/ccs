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
		ID:           id,
		ProjectDir:   "test-project",
		Cwd:          "/tmp/test",
		MessageCount: 1,
		CreatedAt:    now,
		UpdatedAt:    now,
		FileSize:     100,
		FileModTime:  now,
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

func TestTagLifecycle_AddGetRemove(t *testing.T) {
	d := setupTestDB(t)
	insertTestSession(t, d, "sess-1")

	if err := d.AddTag("sess-1", "bug"); err != nil {
		t.Fatalf("AddTag: %v", err)
	}
	// Re-adding the same tag is a no-op (INSERT OR IGNORE).
	if err := d.AddTag("sess-1", "bug"); err != nil {
		t.Fatalf("AddTag duplicate: %v", err)
	}
	if err := d.AddTag("sess-1", "wip"); err != nil {
		t.Fatalf("AddTag second: %v", err)
	}

	tags, err := d.GetSessionTags("sess-1")
	if err != nil {
		t.Fatalf("GetSessionTags: %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("want 2 tags, got %d: %+v", len(tags), tags)
	}

	if err := d.RemoveTag("sess-1", "wip"); err != nil {
		t.Fatalf("RemoveTag: %v", err)
	}
	tags, _ = d.GetSessionTags("sess-1")
	if len(tags) != 1 || tags[0].Name != "bug" {
		t.Errorf("after remove, want [bug], got %+v", tags)
	}
}

func TestGetTagsForSessions_BulkLoad(t *testing.T) {
	d := setupTestDB(t)
	for _, id := range []string{"s1", "s2", "s3"} {
		insertTestSession(t, d, id)
	}
	_ = d.AddTag("s1", "alpha")
	_ = d.AddTag("s1", "beta")
	_ = d.AddTag("s3", "alpha")
	// s2 has no tags — should be omitted from the map.

	tagsByID, err := d.GetTagsForSessions([]string{"s1", "s2", "s3"})
	if err != nil {
		t.Fatalf("GetTagsForSessions: %v", err)
	}
	if len(tagsByID["s1"]) != 2 {
		t.Errorf("s1 want 2 tags, got %+v", tagsByID["s1"])
	}
	if _, ok := tagsByID["s2"]; ok {
		t.Errorf("s2 has no tags, should be absent from map")
	}
	if len(tagsByID["s3"]) != 1 || tagsByID["s3"][0].Name != "alpha" {
		t.Errorf("s3 want [alpha], got %+v", tagsByID["s3"])
	}
}

func TestGetTagsForSessions_EmptyInput(t *testing.T) {
	d := setupTestDB(t)
	got, err := d.GetTagsForSessions(nil)
	if err != nil {
		t.Fatalf("GetTagsForSessions(nil): %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty map, got %+v", got)
	}
}

func TestListSessions_FilterByProjectAndTag(t *testing.T) {
	d := setupTestDB(t)
	insertTestSession(t, d, "a")
	insertTestSession(t, d, "b")
	// Insert one into a different project.
	now := time.Now()
	if err := d.UpsertSession(&model.Session{
		ID: "c", ProjectDir: "other", Cwd: "/x", MessageCount: 1,
		CreatedAt: now, UpdatedAt: now, FileSize: 1, FileModTime: now,
	}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	_ = d.AddTag("a", "keep")
	_ = d.AddTag("b", "keep")

	// Filter by project only.
	got, err := d.ListSessions(model.SessionFilter{ProjectDir: "test-project"})
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("project filter: want 2, got %d", len(got))
	}

	// Filter by tag only.
	got, err = d.ListSessions(model.SessionFilter{Tags: []string{"keep"}})
	if err != nil {
		t.Fatalf("ListSessions by tag: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("tag filter: want 2, got %d", len(got))
	}

	// Tag that nothing has → empty.
	got, _ = d.ListSessions(model.SessionFilter{Tags: []string{"ghost"}})
	if len(got) != 0 {
		t.Errorf("missing tag: want 0, got %d", len(got))
	}
}

func TestListSessions_TagANDSemantics(t *testing.T) {
	d := setupTestDB(t)
	insertTestSession(t, d, "a")
	insertTestSession(t, d, "b")
	_ = d.AddTag("a", "red")
	_ = d.AddTag("a", "blue")
	_ = d.AddTag("b", "red")
	// Only 'a' has BOTH red and blue.

	got, err := d.ListSessions(model.SessionFilter{Tags: []string{"red", "blue"}})
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(got) != 1 || got[0].ID != "a" {
		t.Errorf("AND semantics broken: want just [a], got %+v", got)
	}
}

func TestSearchSessions_MatchesTagName(t *testing.T) {
	d := setupTestDB(t)
	insertTestSession(t, d, "sess-x")
	_ = d.AddTag("sess-x", "needle-tag")

	got, err := d.SearchSessions("needle-tag")
	if err != nil {
		t.Fatalf("SearchSessions: %v", err)
	}
	if len(got) != 1 || got[0].ID != "sess-x" {
		t.Errorf("want [sess-x] matched by tag, got %+v", got)
	}
}

func TestConfig_SetGetDelete(t *testing.T) {
	d := setupTestDB(t)

	// Unset → empty string, no error.
	v, err := d.GetConfig("k")
	if err != nil || v != "" {
		t.Errorf("unset: got %q, err %v", v, err)
	}

	if err := d.SetConfig("k", "v1"); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}
	v, _ = d.GetConfig("k")
	if v != "v1" {
		t.Errorf("after set: got %q, want v1", v)
	}

	// Overwrite same key.
	if err := d.SetConfig("k", "v2"); err != nil {
		t.Fatalf("SetConfig overwrite: %v", err)
	}
	v, _ = d.GetConfig("k")
	if v != "v2" {
		t.Errorf("after overwrite: got %q, want v2", v)
	}

	if err := d.DeleteConfig("k"); err != nil {
		t.Fatalf("DeleteConfig: %v", err)
	}
	v, _ = d.GetConfig("k")
	if v != "" {
		t.Errorf("after delete: got %q, want empty", v)
	}
}

func TestListProjectsAndTagsWithCounts(t *testing.T) {
	d := setupTestDB(t)
	insertTestSession(t, d, "a")
	insertTestSession(t, d, "b")
	_ = d.AddTag("a", "red")
	_ = d.AddTag("b", "red")
	_ = d.AddTag("a", "blue")

	projects, err := d.ListProjectsWithCounts()
	if err != nil {
		t.Fatalf("ListProjectsWithCounts: %v", err)
	}
	if len(projects) != 1 || projects[0].Count != 2 {
		t.Errorf("want 1 project with 2 sessions, got %+v", projects)
	}

	tags, err := d.ListTagsWithCounts()
	if err != nil {
		t.Fatalf("ListTagsWithCounts: %v", err)
	}
	counts := map[string]int{}
	for _, tc := range tags {
		counts[tc.Name] = tc.Count
	}
	if counts["red"] != 2 || counts["blue"] != 1 {
		t.Errorf("want red=2, blue=1, got %+v", counts)
	}
}

func TestDeleteSession_CascadesTags(t *testing.T) {
	d := setupTestDB(t)
	insertTestSession(t, d, "doomed")
	_ = d.AddTag("doomed", "urgent")

	if err := d.DeleteSession("doomed"); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}
	got, _ := d.GetSessionByID("doomed")
	if got != nil {
		t.Error("session not deleted")
	}
	// session_tags rows should be gone too (FK ON DELETE CASCADE).
	tagsByID, _ := d.GetTagsForSessions([]string{"doomed"})
	if len(tagsByID["doomed"]) != 0 {
		t.Errorf("tag association survived delete: %+v", tagsByID["doomed"])
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
