package indexer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/mihai/ccs/internal/format"
)

// ParsedSession holds extracted metadata from a JSONL session file.
type ParsedSession struct {
	SessionID    string
	Cwd          string
	GitBranch    string
	Name         string
	FirstMessage string
	LastMessage  string
	MessageCount int
	FirstTime    time.Time
	LastTime     time.Time
}

// JsonlEntry represents one line in a CC session JSONL file.
type JsonlEntry struct {
	SessionID string          `json:"sessionId"`
	Cwd       string          `json:"cwd"`
	Timestamp string          `json:"timestamp"`
	GitBranch string          `json:"gitBranch"`
	Slug      string          `json:"slug"`
	AgentName string          `json:"agentName"`
	Type      string          `json:"type"`
	Message   json.RawMessage `json:"message"`
}

// MessageObj represents a parsed message with role and content.
type MessageObj struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// ParseSessionFile reads a JSONL file and extracts session metadata.
func ParseSessionFile(path string) (*ParsedSession, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open session file: %w", err)
	}
	defer f.Close()

	p := &ParsedSession{}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024) // up to 10MB per line

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry JsonlEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue // skip malformed lines
		}

		// Extract session-level fields from the first entry that has them.
		if p.SessionID == "" && entry.SessionID != "" {
			p.SessionID = entry.SessionID
		}
		if p.Cwd == "" && entry.Cwd != "" {
			p.Cwd = entry.Cwd
		}
		if p.GitBranch == "" && entry.GitBranch != "" {
			p.GitBranch = entry.GitBranch
		}
		// Prefer user-set agentName (from --name) over auto-generated slug.
		if entry.AgentName != "" {
			p.Name = entry.AgentName
		} else if p.Name == "" && entry.Slug != "" {
			p.Name = entry.Slug
		}

		// Parse timestamp.
		var ts time.Time
		if entry.Timestamp != "" {
			parsed := format.ParseTime(entry.Timestamp)
			if !parsed.IsZero() {
				ts = parsed
				if p.FirstTime.IsZero() {
					p.FirstTime = ts
				}
			}
		}

		// Only count user/assistant messages.
		if entry.Type != "user" && entry.Type != "assistant" {
			continue
		}

		// Track LastTime only from user/assistant entries so that
		// permission-mode / file-history-snapshot entries don't inflate it.
		if !ts.IsZero() {
			p.LastTime = ts
		}

		if len(entry.Message) == 0 {
			continue
		}

		var msg MessageObj
		if err := json.Unmarshal(entry.Message, &msg); err != nil {
			continue
		}

		if msg.Role != "user" && msg.Role != "assistant" {
			continue
		}

		content := ExtractContent(msg.Content)
		if content == "" {
			continue
		}

		p.MessageCount++

		if p.FirstMessage == "" {
			p.FirstMessage = truncate(content, 200)
		}
		p.LastMessage = truncate(content, 200)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan session file: %w", err)
	}

	if p.SessionID == "" {
		return nil, nil // empty or unparseable file
	}

	return p, nil
}

// ExtractContent handles message.content being either a string or an array
// of content blocks. For arrays, it concatenates text from blocks with type "text".
func ExtractContent(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	// Try string first.
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}

	// Try array of content blocks.
	var blocks []json.RawMessage
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return ""
	}

	var result string
	for _, block := range blocks {
		var cb struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}
		if err := json.Unmarshal(block, &cb); err != nil {
			continue
		}
		if cb.Type == "text" && cb.Text != "" {
			if result != "" {
				result += " "
			}
			result += cb.Text
		}
	}
	return result
}

// truncate returns the first n characters of s (rune-aware).
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}
