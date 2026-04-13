package indexer

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestScanSessions_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	// No projects/ dir at all.
	files, partial, err := ScanSessions(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if partial {
		t.Error("expected partial=false for nonexistent projects dir")
	}
	if len(files) != 0 {
		t.Errorf("expected no files, got %d", len(files))
	}
}

func TestScanSessions_FindsJSONLFiles(t *testing.T) {
	dir := t.TempDir()
	projDir := filepath.Join(dir, "projects", "my-project")
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create two session files.
	for _, name := range []string{"abc-123.jsonl", "def-456.jsonl"} {
		if err := os.WriteFile(filepath.Join(projDir, name), []byte(`{}`+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Create a non-JSONL file that should be ignored.
	if err := os.WriteFile(filepath.Join(projDir, "notes.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	files, partial, err := ScanSessions(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if partial {
		t.Error("expected partial=false for clean scan")
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}

	ids := map[string]bool{}
	for _, f := range files {
		ids[f.SessionID] = true
	}
	if !ids["abc-123"] || !ids["def-456"] {
		t.Errorf("unexpected session IDs: %v", ids)
	}
}

func TestScanSessions_InaccessibleSubdirectory_ReturnsPartial(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod not effective on Windows")
	}
	if os.Getuid() == 0 {
		t.Skip("test requires non-root to enforce permission denial")
	}

	dir := t.TempDir()
	projDir := filepath.Join(dir, "projects")

	// Create an accessible project with a session file.
	accessibleDir := filepath.Join(projDir, "accessible-project")
	if err := os.MkdirAll(accessibleDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(accessibleDir, "good-session.jsonl"), []byte(`{}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create an inaccessible project directory.
	inaccessibleDir := filepath.Join(projDir, "locked-project")
	if err := os.MkdirAll(inaccessibleDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inaccessibleDir, "hidden-session.jsonl"), []byte(`{}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Remove read+execute permission so WalkDir can't enter it.
	if err := os.Chmod(inaccessibleDir, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.Chmod(inaccessibleDir, 0o755) // restore for cleanup
	})

	files, partial, err := ScanSessions(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !partial {
		t.Error("expected partial=true when a subdirectory is inaccessible")
	}

	// The accessible session should still be found.
	if len(files) != 1 {
		t.Fatalf("expected 1 file from accessible dir, got %d", len(files))
	}
	if files[0].SessionID != "good-session" {
		t.Errorf("expected session ID 'good-session', got %q", files[0].SessionID)
	}
}
