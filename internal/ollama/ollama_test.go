package ollama

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClient_Defaults(t *testing.T) {
	c := NewClient("", "")
	if c.BaseURL != DefaultURL {
		t.Errorf("BaseURL = %q, want %q", c.BaseURL, DefaultURL)
	}
	if c.Model != DefaultModel {
		t.Errorf("Model = %q, want %q", c.Model, DefaultModel)
	}
}

func TestNewClient_Overrides(t *testing.T) {
	c := NewClient("http://example.com", "llama3.2")
	if c.BaseURL != "http://example.com" || c.Model != "llama3.2" {
		t.Errorf("overrides not applied: %+v", c)
	}
}

func TestGenerate_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"prompt":"please summarize"`) {
			t.Errorf("request body missing prompt: %s", body)
		}
		json.NewEncoder(w).Encode(generateResponse{Response: "a concise summary", Done: true})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-model")
	got, err := c.Generate("please summarize")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if got != "a concise summary" {
		t.Errorf("got %q, want %q", got, "a concise summary")
	}
}

func TestGenerate_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "model not found", http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "missing-model")
	_, err := c.Generate("hi")
	if err == nil {
		t.Fatal("expected error on 404")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error %q should mention status", err)
	}
}

func TestPing_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "")
	if err := c.Ping(); err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestPing_ConnectionRefused(t *testing.T) {
	// Point at a port nothing is listening on.
	c := NewClient("http://127.0.0.1:1", "")
	err := c.Ping()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "cannot reach Ollama") {
		t.Errorf("error %q should mention Ollama reachability", err)
	}
}

func TestPing_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "")
	err := c.Ping()
	if err == nil || !strings.Contains(err.Error(), "500") {
		t.Errorf("expected 500 error, got %v", err)
	}
}
