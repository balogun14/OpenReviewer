package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAICompatibleProviderBuildsRequestAndParsesFindings(t *testing.T) {
	var received chatCompletionRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("unexpected authorization header: %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("X-Test") != "yes" {
			t.Fatalf("expected custom header")
		}

		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices": [
				{
					"message": {
						"role": "assistant",
						"content": "{\"findings\":[{\"severity\":\"high\",\"category\":\"security\",\"subcategory\":\"secrets\",\"title\":\"Secret in diff\",\"description\":\"A secret appears in the diff.\",\"impact\":\"Credential exposure\",\"recommendation\":\"Move it to a secret store.\",\"file\":\"main.go\",\"line\":12,\"confidence\":\"high\"}],\"summary\":\"one finding\",\"recommendation\":\"request_changes\"}"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	provider := NewOpenAICompatibleProvider(OpenAICompatibleConfig{
		BaseURL: server.URL + "/v1",
		APIKey:  "test-key",
		Model:   "test-model",
		Headers: map[string]string{"X-Test": "yes"},
	})

	findings, err := provider.Review(context.Background(), Request{
		PromptID:  "security.secrets",
		PersonaID: "security-engineer",
		Prompt:    "review this diff",
	})
	if err != nil {
		t.Fatalf("Review returned error: %v", err)
	}

	if received.Model != "test-model" {
		t.Fatalf("unexpected model: %q", received.Model)
	}
	if received.ResponseFormat.Type != "json_object" {
		t.Fatalf("unexpected response format: %q", received.ResponseFormat.Type)
	}
	if len(received.Messages) != 2 {
		t.Fatalf("expected two messages, got %d", len(received.Messages))
	}
	if received.Messages[1].Role != "user" || !strings.Contains(received.Messages[1].Content, "review this diff") {
		t.Fatalf("unexpected user message: %+v", received.Messages[1])
	}

	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %d", len(findings))
	}
	if findings[0].Persona != "security-engineer" {
		t.Fatalf("unexpected persona: %q", findings[0].Persona)
	}
	if findings[0].Location.File != "main.go" {
		t.Fatalf("unexpected finding file: %q", findings[0].Location.File)
	}
}

func TestOpenAICompatibleProviderMarksRateLimitRetryable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "slow down", http.StatusTooManyRequests)
	}))
	defer server.Close()

	provider := NewOpenAICompatibleProvider(OpenAICompatibleConfig{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Model:   "test-model",
	})

	_, err := provider.Review(context.Background(), Request{Prompt: "review"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !isRetryable(err) {
		t.Fatalf("expected retryable error, got %T: %v", err, err)
	}
}
