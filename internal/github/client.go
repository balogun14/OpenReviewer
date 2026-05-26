package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const DefaultAPIBaseURL = "https://api.github.com"

type Client struct {
	baseURL       string
	appID         string
	privateKeyPEM []byte
	httpClient    *http.Client
	now           func() time.Time
}

type ClientConfig struct {
	BaseURL       string
	AppID         string
	PrivateKeyPEM []byte
	HTTPClient    *http.Client
}

func NewClient(config ClientConfig) *Client {
	baseURL := strings.TrimRight(config.BaseURL, "/")
	if baseURL == "" {
		baseURL = DefaultAPIBaseURL
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	return &Client{
		baseURL:       baseURL,
		appID:         config.AppID,
		privateKeyPEM: config.PrivateKeyPEM,
		httpClient:    httpClient,
		now:           time.Now,
	}
}

type PullRequestFile struct {
	Filename string `json:"filename"`
	Status   string `json:"status"`
	Patch    string `json:"patch"`
	Changes  int    `json:"changes"`
}

type PullRequestMetadata struct {
	HeadSHA string `json:"head_sha"`
	BaseSHA string `json:"base_sha"`
	Title   string `json:"title"`
	Body    string `json:"body"`
}

type CheckRunConclusion string

const (
	CheckRunSuccess CheckRunConclusion = "success"
	CheckRunNeutral CheckRunConclusion = "neutral"
	CheckRunFailure CheckRunConclusion = "failure"
)

func (c *Client) InstallationToken(ctx context.Context, installationID int64) (string, error) {
	if installationID <= 0 {
		return "", fmt.Errorf("installation ID is required")
	}

	jwt, err := AppJWT(c.appID, c.privateKeyPEM, c.now())
	if err != nil {
		return "", err
	}

	var response struct {
		Token string `json:"token"`
	}
	path := "/app/installations/" + strconv.FormatInt(installationID, 10) + "/access_tokens"
	if err := c.do(ctx, http.MethodPost, path, jwt, nil, &response); err != nil {
		return "", err
	}
	if strings.TrimSpace(response.Token) == "" {
		return "", fmt.Errorf("GitHub installation token response did not include token")
	}

	return response.Token, nil
}

func (c *Client) PullRequest(ctx context.Context, token string, owner string, repo string, number int) (PullRequestMetadata, error) {
	var response struct {
		Title string `json:"title"`
		Body  string `json:"body"`
		Head  struct {
			SHA string `json:"sha"`
		} `json:"head"`
		Base struct {
			SHA string `json:"sha"`
		} `json:"base"`
	}

	path := repoPath(owner, repo, fmt.Sprintf("/pulls/%d", number))
	if err := c.do(ctx, http.MethodGet, path, token, nil, &response); err != nil {
		return PullRequestMetadata{}, err
	}

	return PullRequestMetadata{
		HeadSHA: response.Head.SHA,
		BaseSHA: response.Base.SHA,
		Title:   response.Title,
		Body:    response.Body,
	}, nil
}

func (c *Client) PullRequestFiles(ctx context.Context, token string, owner string, repo string, number int) ([]PullRequestFile, error) {
	var files []PullRequestFile

	for page := 1; ; page++ {
		var pageFiles []PullRequestFile
		path := repoPath(owner, repo, fmt.Sprintf("/pulls/%d/files?per_page=100&page=%d", number, page))
		if err := c.do(ctx, http.MethodGet, path, token, nil, &pageFiles); err != nil {
			return nil, err
		}

		files = append(files, pageFiles...)
		if len(pageFiles) < 100 {
			break
		}
	}

	return files, nil
}

func (c *Client) CreateIssueComment(ctx context.Context, token string, owner string, repo string, number int, body string) error {
	payload := map[string]string{"body": body}
	path := repoPath(owner, repo, fmt.Sprintf("/issues/%d/comments", number))
	return c.do(ctx, http.MethodPost, path, token, payload, nil)
}

func (c *Client) CreateReviewComment(ctx context.Context, token string, owner string, repo string, number int, commitID string, path string, line int, body string) error {
	payload := map[string]any{
		"body":      body,
		"commit_id": commitID,
		"path":      path,
		"line":      line,
		"side":      "RIGHT",
	}
	apiPath := repoPath(owner, repo, fmt.Sprintf("/pulls/%d/comments", number))
	return c.do(ctx, http.MethodPost, apiPath, token, payload, nil)
}

func (c *Client) CreateCheckRun(ctx context.Context, token string, owner string, repo string, headSHA string, conclusion CheckRunConclusion, summary string) error {
	payload := map[string]any{
		"name":       "OpenReview AI",
		"head_sha":   headSHA,
		"status":     "completed",
		"conclusion": conclusion,
		"output": map[string]string{
			"title":   "OpenReview AI",
			"summary": summary,
		},
	}

	path := repoPath(owner, repo, "/check-runs")
	return c.do(ctx, http.MethodPost, path, token, payload, nil)
}

func (c *Client) do(ctx context.Context, method string, path string, token string, payload any, target any) error {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("encode GitHub request: %w", err)
		}
		body = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("create GitHub request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if strings.TrimSpace(token) != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send GitHub request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return fmt.Errorf("read GitHub response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("GitHub returned status %d: %s", resp.StatusCode, string(respBody))
	}

	if target == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, target); err != nil {
		return fmt.Errorf("decode GitHub response: %w", err)
	}

	return nil
}

func repoPath(owner string, repo string, suffix string) string {
	return "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(repo) + suffix
}
