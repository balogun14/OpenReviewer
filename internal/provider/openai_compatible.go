package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/openreview-ai/openreview/internal/finding"
	"github.com/openreview-ai/openreview/internal/llm"
)

type OpenAICompatibleConfig struct {
	BaseURL string
	APIKey  string
	Model   string
	Headers map[string]string
}

type OpenAICompatibleProvider struct {
	baseURL string
	apiKey  string
	model   string
	headers map[string]string
	client  *http.Client
}

func NewOpenAICompatibleProvider(config OpenAICompatibleConfig) OpenAICompatibleProvider {
	client := &http.Client{Timeout: 60 * time.Second}
	return NewOpenAICompatibleProviderWithClient(config, client)
}

func NewOpenAICompatibleProviderWithClient(config OpenAICompatibleConfig, client *http.Client) OpenAICompatibleProvider {
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}

	return OpenAICompatibleProvider{
		baseURL: strings.TrimRight(config.BaseURL, "/"),
		apiKey:  config.APIKey,
		model:   config.Model,
		headers: config.Headers,
		client:  client,
	}
}

func (p OpenAICompatibleProvider) Review(ctx context.Context, req Request) ([]finding.Finding, error) {
	if p.baseURL == "" {
		return nil, fmt.Errorf("provider base URL is required")
	}
	if p.model == "" {
		return nil, fmt.Errorf("provider model is required")
	}
	if p.apiKey == "" {
		return nil, fmt.Errorf("provider API key is required")
	}

	payload := chatCompletionRequest{
		Model: p.model,
		Messages: []chatMessage{
			{
				Role:    "system",
				Content: "You are OpenReview AI. Return only valid JSON matching the requested output contract.",
			},
			{
				Role:    "user",
				Content: req.Prompt,
			},
		},
		ResponseFormat: responseFormat{Type: "json_object"},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode provider request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create provider request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	for key, value := range p.headers {
		if strings.TrimSpace(value) != "" {
			httpReq.Header.Set(key, value)
		}
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, RetryableError{Err: fmt.Errorf("send provider request: %w", err)}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("read provider response: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, RetryableError{Err: fmt.Errorf("provider returned retryable status %d: %s", resp.StatusCode, string(respBody))}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("provider returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var completion chatCompletionResponse
	if err := json.Unmarshal(respBody, &completion); err != nil {
		return nil, fmt.Errorf("decode provider response: %w", err)
	}
	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("provider response did not include choices")
	}

	parsed, err := llm.ParseReviewResponse(completion.Choices[0].Message.Content, llm.ResponseContext{
		PromptID:  req.PromptID,
		PersonaID: req.PersonaID,
	})
	if err != nil {
		return nil, fmt.Errorf("parse provider review response: %w", err)
	}

	return parsed.Findings, nil
}

type chatCompletionRequest struct {
	Model          string         `json:"model"`
	Messages       []chatMessage  `json:"messages"`
	ResponseFormat responseFormat `json:"response_format"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}
