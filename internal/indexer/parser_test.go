package indexer

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempJSONL(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseSessionFile_Normal(t *testing.T) {
	dir := t.TempDir()
	path := writeTempJSONL(t, dir, "abc-123.jsonl", `{"sessionId":"abc-123","cwd":"/home/user/project","timestamp":"2026-01-15T10:00:00Z","gitBranch":"main","slug":"my-session","type":"user","message":{"role":"user","content":"Hello, can you help me?"}}
{"sessionId":"abc-123","cwd":"/home/user/project","timestamp":"2026-01-15T10:01:00Z","type":"assistant","message":{"role":"assistant","content":"Sure, I can help you with that."}}
{"sessionId":"abc-123","cwd":"/home/user/project","timestamp":"2026-01-15T10:02:00Z","type":"user","message":{"role":"user","content":"Thanks, please refactor this function."}}
`)

	parsed, err := ParseSessionFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed == nil {
		t.Fatal("expected parsed result, got nil")
	}

	if parsed.SessionID != "abc-123" {
		t.Errorf("SessionID = %q, want %q", parsed.SessionID, "abc-123")
	}
	if parsed.Cwd != "/home/user/project" {
		t.Errorf("Cwd = %q, want %q", parsed.Cwd, "/home/user/project")
	}
	if parsed.GitBranch != "main" {
		t.Errorf("GitBranch = %q, want %q", parsed.GitBranch, "main")
	}
	if parsed.Name != "my-session" {
		t.Errorf("Name = %q, want %q", parsed.Name, "my-session")
	}
	if parsed.MessageCount != 3 {
		t.Errorf("MessageCount = %d, want %d", parsed.MessageCount, 3)
	}
	if parsed.FirstMessage != "Hello, can you help me?" {
		t.Errorf("FirstMessage = %q, want %q", parsed.FirstMessage, "Hello, can you help me?")
	}
	if parsed.LastMessage != "Thanks, please refactor this function." {
		t.Errorf("LastMessage = %q, want %q", parsed.LastMessage, "Thanks, please refactor this function.")
	}
	if parsed.FirstTime.IsZero() {
		t.Error("FirstTime is zero")
	}
	if parsed.LastTime.IsZero() {
		t.Error("LastTime is zero")
	}
	if !parsed.LastTime.After(parsed.FirstTime) {
		t.Error("LastTime should be after FirstTime")
	}
}

func TestParseSessionFile_ArrayContentBlocks(t *testing.T) {
	dir := t.TempDir()
	path := writeTempJSONL(t, dir, "def-456.jsonl", `{"sessionId":"def-456","cwd":"/tmp","timestamp":"2026-02-01T08:00:00Z","type":"user","message":{"role":"user","content":[{"type":"text","text":"First block"},{"type":"image","url":"http://example.com/img.png"},{"type":"text","text":"second block"}]}}
{"sessionId":"def-456","cwd":"/tmp","timestamp":"2026-02-01T08:01:00Z","type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"I see the image and text."}]}}
`)

	parsed, err := ParseSessionFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed == nil {
		t.Fatal("expected parsed result, got nil")
	}

	if parsed.SessionID != "def-456" {
		t.Errorf("SessionID = %q, want %q", parsed.SessionID, "def-456")
	}
	if parsed.MessageCount != 2 {
		t.Errorf("MessageCount = %d, want %d", parsed.MessageCount, 2)
	}
	if parsed.FirstMessage != "First block second block" {
		t.Errorf("FirstMessage = %q, want %q", parsed.FirstMessage, "First block second block")
	}
	if parsed.LastMessage != "I see the image and text." {
		t.Errorf("LastMessage = %q, want %q", parsed.LastMessage, "I see the image and text.")
	}
}

func TestParseSessionFile_MissingFields(t *testing.T) {
	dir := t.TempDir()
	// No gitBranch, no slug, minimal fields
	path := writeTempJSONL(t, dir, "ghi-789.jsonl", `{"sessionId":"ghi-789","cwd":"/opt","timestamp":"2026-03-01T12:00:00Z","type":"user","message":{"role":"user","content":"just a question"}}
`)

	parsed, err := ParseSessionFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed == nil {
		t.Fatal("expected parsed result, got nil")
	}

	if parsed.SessionID != "ghi-789" {
		t.Errorf("SessionID = %q, want %q", parsed.SessionID, "ghi-789")
	}
	if parsed.GitBranch != "" {
		t.Errorf("GitBranch = %q, want empty", parsed.GitBranch)
	}
	if parsed.Name != "" {
		t.Errorf("Name = %q, want empty", parsed.Name)
	}
	if parsed.MessageCount != 1 {
		t.Errorf("MessageCount = %d, want %d", parsed.MessageCount, 1)
	}
}

func TestParseSessionFile_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := writeTempJSONL(t, dir, "empty.jsonl", "")

	parsed, err := ParseSessionFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed != nil {
		t.Errorf("expected nil for empty file, got %+v", parsed)
	}
}

func TestParseSessionFile_SkipsNonMessageTypes(t *testing.T) {
	dir := t.TempDir()
	path := writeTempJSONL(t, dir, "mixed.jsonl", `{"sessionId":"mix-001","cwd":"/home","timestamp":"2026-04-01T09:00:00Z","type":"permission-mode","message":{"role":"user","content":"allow all"}}
{"sessionId":"mix-001","cwd":"/home","timestamp":"2026-04-01T09:01:00Z","type":"file-history-snapshot","message":{"role":"assistant","content":"snapshot data"}}
{"sessionId":"mix-001","cwd":"/home","timestamp":"2026-04-01T09:02:00Z","type":"user","message":{"role":"user","content":"real user message"}}
{"sessionId":"mix-001","cwd":"/home","timestamp":"2026-04-01T09:03:00Z","type":"assistant","message":{"role":"assistant","content":"real assistant reply"}}
`)

	parsed, err := ParseSessionFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed == nil {
		t.Fatal("expected parsed result, got nil")
	}

	if parsed.MessageCount != 2 {
		t.Errorf("MessageCount = %d, want 2 (should skip non user/assistant types)", parsed.MessageCount)
	}
	if parsed.FirstMessage != "real user message" {
		t.Errorf("FirstMessage = %q, want %q", parsed.FirstMessage, "real user message")
	}
}

func TestParseSessionFile_TruncatesLongMessages(t *testing.T) {
	dir := t.TempDir()
	longMsg := ""
	for i := 0; i < 250; i++ {
		longMsg += "x"
	}
	path := writeTempJSONL(t, dir, "long.jsonl", `{"sessionId":"long-001","cwd":"/tmp","timestamp":"2026-04-01T10:00:00Z","type":"user","message":{"role":"user","content":"`+longMsg+`"}}
`)

	parsed, err := ParseSessionFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed == nil {
		t.Fatal("expected parsed result, got nil")
	}

	if len([]rune(parsed.FirstMessage)) != 200 {
		t.Errorf("FirstMessage length = %d, want 200", len([]rune(parsed.FirstMessage)))
	}
}

func TestScanSessions_SkipsSubagents(t *testing.T) {
	dir := t.TempDir()
	claudeDir := dir

	// Create projects directory structure
	projectDir := filepath.Join(claudeDir, "projects", "my-project")
	subagentDir := filepath.Join(projectDir, "abc-session", "subagents")

	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(subagentDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a normal session file
	writeTempJSONL(t, projectDir, "session-001.jsonl", `{"sessionId":"session-001"}`)

	// Create a subagent file that should be skipped
	writeTempJSONL(t, subagentDir, "sub-001.jsonl", `{"sessionId":"sub-001"}`)

	files, _, err := ScanSessions(claudeDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].SessionID != "session-001" {
		t.Errorf("SessionID = %q, want %q", files[0].SessionID, "session-001")
	}
	if files[0].ProjectDir != "my-project" {
		t.Errorf("ProjectDir = %q, want %q", files[0].ProjectDir, "my-project")
	}
}

func TestScanSessions_NonexistentDir(t *testing.T) {
	files, _, err := ScanSessions("/nonexistent/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if files != nil {
		t.Errorf("expected nil, got %v", files)
	}
}
