package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const DefaultURL = "http://localhost:11434"
const DefaultModel = "gemma3:4b"

type Client struct {
	BaseURL string
	Model   string
	client  *http.Client
}

func NewClient(baseURL, model string) *Client {
	if baseURL == "" {
		baseURL = DefaultURL
	}
	if model == "" {
		model = DefaultModel
	}
	return &Client{
		BaseURL: baseURL,
		Model:   model,
		client:  &http.Client{Timeout: 120 * time.Second},
	}
}

type generateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type generateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// Generate sends a prompt to Ollama and returns the full response text.
func (c *Client) Generate(prompt string) (string, error) {
	body, err := json.Marshal(generateRequest{
		Model:  c.Model,
		Prompt: prompt,
		Stream: false,
	})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.client.Post(c.BaseURL+"/api/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ollama request failed (is Ollama running?): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode ollama response: %w", err)
	}

	return result.Response, nil
}

// Ping checks if Ollama is reachable.
func (c *Client) Ping() error {
	resp, err := c.client.Get(c.BaseURL + "/api/tags")
	if err != nil {
		return fmt.Errorf("cannot reach Ollama at %s: %w", c.BaseURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama returned status %d at %s", resp.StatusCode, c.BaseURL)
	}
	return nil
}
