package summarizer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mihai/ccs/internal/format"
	"github.com/mihai/ccs/internal/indexer"
	"github.com/mihai/ccs/internal/ollama"
)

const (
	headMessages = 5
	midMessages  = 3
	tailMessages = 5
	maxMsgLen    = 300
)

// Summarize reads a session JSONL file, extracts key messages, and generates
// a summary via the given Ollama client.
func Summarize(client *ollama.Client, jsonlPath string) (string, error) {
	messages, err := extractMessages(jsonlPath)
	if err != nil {
		return "", fmt.Errorf("extract messages: %w", err)
	}

	if len(messages) == 0 {
		return "", fmt.Errorf("no messages found in session")
	}

	excerpt := buildExcerpt(messages)
	prompt := buildPrompt(excerpt)

	summary, err := client.Generate(prompt)
	if err != nil {
		return "", fmt.Errorf("generate summary: %w", err)
	}

	return strings.TrimSpace(summary), nil
}

type message struct {
	Role    string
	Content string
}

func extractMessages(path string) ([]message, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var messages []message
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry indexer.JsonlEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}

		if entry.Type != "user" && entry.Type != "assistant" {
			continue
		}
		if len(entry.Message) == 0 {
			continue
		}

		var msg indexer.MessageObj
		if err := json.Unmarshal(entry.Message, &msg); err != nil {
			continue
		}
		if msg.Role != "user" && msg.Role != "assistant" {
			continue
		}

		content := indexer.ExtractContent(msg.Content)
		if content == "" {
			continue
		}

		messages = append(messages, message{Role: msg.Role, Content: content})
	}

	return messages, scanner.Err()
}

func buildExcerpt(messages []message) []message {
	n := len(messages)

	if n <= headMessages+tailMessages {
		return messages
	}

	var excerpt []message

	// Head
	for i := 0; i < headMessages && i < n; i++ {
		excerpt = append(excerpt, messages[i])
	}

	// Mid (from the middle of the conversation)
	if n > headMessages+midMessages+tailMessages {
		midStart := n / 2
		for i := 0; i < midMessages && midStart+i < n; i++ {
			excerpt = append(excerpt, messages[midStart+i])
		}
	}

	// Tail
	tailStart := n - tailMessages
	if tailStart < 0 {
		tailStart = 0
	}
	for i := tailStart; i < n; i++ {
		excerpt = append(excerpt, messages[i])
	}

	return excerpt
}

func buildPrompt(messages []message) string {
	var sb strings.Builder
	sb.WriteString("Summarize the following Claude Code conversation in 2-3 sentences. ")
	sb.WriteString("Focus on: what the user was trying to accomplish, what was done, and the outcome. ")
	sb.WriteString("Be concise and factual.\n\n")
	sb.WriteString("--- CONVERSATION EXCERPT ---\n\n")

	for _, m := range messages {
		content := format.Truncate(m.Content, maxMsgLen)
		sb.WriteString(fmt.Sprintf("[%s]: %s\n\n", m.Role, content))
	}

	sb.WriteString("--- END EXCERPT ---\n\n")
	sb.WriteString("Summary:")
	return sb.String()
}

