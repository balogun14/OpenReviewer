package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openreview-ai/openreview/internal/github"
	"github.com/openreview-ai/openreview/internal/prompt"
	"github.com/openreview-ai/openreview/internal/provider"
	"github.com/openreview-ai/openreview/internal/review"
)

func TestGitHubWebhookReviewsPullRequestAndPublishesSummary(t *testing.T) {
	gh := &fakeGitHubClient{
		files: []github.PullRequestFile{
			{
				Filename: "main.go",
				Status:   "modified",
				Patch:    "@@ -1,2 +1,2 @@\n package main\n+const password = \"secret\"",
			},
		},
		metadata: github.PullRequestMetadata{
			HeadSHA: "head-sha",
			BaseSHA: "base-sha",
			Title:   "Add secret",
			Body:    "test body",
		},
	}

	renderer := prompt.NewRenderer(prompt.NewLoader("../../prompts"), prompt.DefaultManifest())
	engine := review.NewEngine(provider.NewMockProvider(), renderer, review.DefaultReviewerPersonas())
	server := NewServer(ServerOptions{
		ReviewEngine: engine,
		GitHubClient: gh,
	})

	body := pullRequestPayload(t, "opened")
	req := httptest.NewRequest(http.MethodPost, "/webhooks/github", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-GitHub-Delivery", "delivery-1")
	rec := httptest.NewRecorder()

	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
	if gh.tokenCalls != 1 {
		t.Fatalf("expected token call, got %d", gh.tokenCalls)
	}
	if gh.fileCalls != 1 {
		t.Fatalf("expected file call, got %d", gh.fileCalls)
	}
	if gh.summaryComments != 1 {
		t.Fatalf("expected summary comment, got %d", gh.summaryComments)
	}
	if gh.checkRuns != 1 {
		t.Fatalf("expected check run, got %d", gh.checkRuns)
	}

	dupReq := httptest.NewRequest(http.MethodPost, "/webhooks/github", bytes.NewReader(body))
	dupReq.Header.Set("X-GitHub-Event", "pull_request")
	dupReq.Header.Set("X-GitHub-Delivery", "delivery-1")
	dupRec := httptest.NewRecorder()

	server.Routes().ServeHTTP(dupRec, dupReq)

	if dupRec.Code != http.StatusAccepted {
		t.Fatalf("unexpected duplicate status: %d", dupRec.Code)
	}
	if gh.fileCalls != 1 {
		t.Fatalf("duplicate delivery fetched files again")
	}
}

func pullRequestPayload(t *testing.T, action string) []byte {
	t.Helper()

	payload := map[string]any{
		"action": action,
		"number": 7,
		"installation": map[string]any{
			"id": 99,
		},
		"pull_request": map[string]any{
			"title": "Original title",
			"body":  "Original body",
			"head": map[string]any{
				"sha": "event-head",
			},
		},
		"repository": map[string]any{
			"name":      "api",
			"full_name": "acme/api",
			"owner": map[string]any{
				"login": "acme",
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

type fakeGitHubClient struct {
	tokenCalls      int
	fileCalls       int
	summaryComments int
	inlineComments  int
	checkRuns       int
	metadata        github.PullRequestMetadata
	files           []github.PullRequestFile
}

func (c *fakeGitHubClient) InstallationToken(ctx context.Context, installationID int64) (string, error) {
	c.tokenCalls++
	return "token", nil
}

func (c *fakeGitHubClient) PullRequest(ctx context.Context, token string, owner string, repo string, number int) (github.PullRequestMetadata, error) {
	return c.metadata, nil
}

func (c *fakeGitHubClient) PullRequestFiles(ctx context.Context, token string, owner string, repo string, number int) ([]github.PullRequestFile, error) {
	c.fileCalls++
	return c.files, nil
}

func (c *fakeGitHubClient) CreateIssueComment(ctx context.Context, token string, owner string, repo string, number int, body string) error {
	c.summaryComments++
	return nil
}

func (c *fakeGitHubClient) CreateReviewComment(ctx context.Context, token string, owner string, repo string, number int, commitID string, path string, line int, body string) error {
	c.inlineComments++
	return nil
}

func (c *fakeGitHubClient) CreateCheckRun(ctx context.Context, token string, owner string, repo string, headSHA string, conclusion github.CheckRunConclusion, summary string) error {
	c.checkRuns++
	return nil
}
