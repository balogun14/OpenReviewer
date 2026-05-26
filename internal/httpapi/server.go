package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/openreview-ai/openreview/internal/finding"
	"github.com/openreview-ai/openreview/internal/github"
	"github.com/openreview-ai/openreview/internal/profile"
	"github.com/openreview-ai/openreview/internal/review"
)

type ServerOptions struct {
	Logger              *slog.Logger
	ReviewEngine        *review.Engine
	GitHubWebhookSecret string
	GitHubClient        gitHubClient
}

type gitHubClient interface {
	InstallationToken(ctx context.Context, installationID int64) (string, error)
	PullRequest(ctx context.Context, token string, owner string, repo string, number int) (github.PullRequestMetadata, error)
	PullRequestFiles(ctx context.Context, token string, owner string, repo string, number int) ([]github.PullRequestFile, error)
	CreateIssueComment(ctx context.Context, token string, owner string, repo string, number int, body string) error
	CreateReviewComment(ctx context.Context, token string, owner string, repo string, number int, commitID string, path string, line int, body string) error
	CreateCheckRun(ctx context.Context, token string, owner string, repo string, headSHA string, conclusion github.CheckRunConclusion, summary string) error
}

type Server struct {
	logger              *slog.Logger
	reviewEngine        *review.Engine
	gitHubWebhookSecret string
	gitHubClient        gitHubClient
	mu                  sync.RWMutex
	reviews             map[string]review.Review
	profiles            map[string]profile.ReviewProfile
	repositories        map[string]Repository
	deliveries          map[string]struct{}
}

type Repository struct {
	FullName string `json:"fullName"`
}

func NewServer(opts ServerOptions) *Server {
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	defaultProfile := profile.SecurityFirst()
	return &Server{
		logger:              logger,
		reviewEngine:        opts.ReviewEngine,
		gitHubWebhookSecret: opts.GitHubWebhookSecret,
		gitHubClient:        opts.GitHubClient,
		reviews:             make(map[string]review.Review),
		deliveries:          make(map[string]struct{}),
		profiles: map[string]profile.ReviewProfile{
			defaultProfile.ID: defaultProfile,
		},
		repositories: make(map[string]Repository),
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("POST /webhooks/github", s.githubWebhook)
	mux.HandleFunc("GET /reviews/{id}", s.getReview)
	mux.HandleFunc("GET /repositories", s.listRepositories)
	mux.HandleFunc("POST /review-profiles", s.createReviewProfile)
	mux.HandleFunc("PUT /review-profiles/{id}", s.updateReviewProfile)
	return mux
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) githubWebhook(w http.ResponseWriter, r *http.Request) {
	eventName := r.Header.Get("X-GitHub-Event")
	if eventName != "pull_request" {
		writeJSON(w, http.StatusAccepted, map[string]string{"status": "ignored"})
		return
	}
	if s.seenDelivery(r.Header.Get("X-GitHub-Delivery")) {
		writeJSON(w, http.StatusAccepted, map[string]string{"status": "duplicate"})
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 10<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "read_body_failed")
		return
	}

	if err := github.ValidateSignature(r, body, s.gitHubWebhookSecret); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid_signature")
		return
	}

	event, err := github.DecodePullRequestEvent(body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_payload")
		return
	}

	if !github.IsPullRequestReviewAction(event.Action) {
		writeJSON(w, http.StatusAccepted, map[string]string{"status": "ignored"})
		return
	}

	input, token, metadata, files, err := s.reviewInputFromEvent(r.Context(), event)
	if err != nil {
		s.logger.Error("prepare review failed", "error", err)
		writeError(w, http.StatusInternalServerError, "prepare_review_failed")
		return
	}

	result, err := s.reviewEngine.ReviewPullRequest(r.Context(), input)
	if err != nil {
		s.logger.Error("review failed", "error", err)
		writeError(w, http.StatusInternalServerError, "review_failed")
		return
	}

	if s.gitHubClient != nil && token != "" {
		if err := s.publishReview(r.Context(), token, event, metadata.HeadSHA, files, result); err != nil {
			s.logger.Error("publish review failed", "error", err)
		}
	}

	s.mu.Lock()
	s.reviews[result.ID] = result
	s.repositories[result.Repository] = Repository{FullName: result.Repository}
	s.mu.Unlock()

	writeJSON(w, http.StatusAccepted, map[string]string{
		"status":   "reviewed",
		"reviewId": result.ID,
	})
}

func (s *Server) reviewInputFromEvent(ctx context.Context, event github.PullRequestEvent) (review.PullRequestInput, string, github.PullRequestMetadata, []github.PullRequestFile, error) {
	input := review.PullRequestInput{
		Owner:       event.Repository.Owner.Login,
		Repository:  event.Repository.Name,
		Number:      event.Number,
		Title:       event.PullRequest.Title,
		Description: event.PullRequest.Body,
		Diff:        event.PullRequest.Head.SHA,
		Profile:     s.defaultProfile(),
	}

	if s.gitHubClient == nil {
		return input, "", github.PullRequestMetadata{HeadSHA: event.PullRequest.Head.SHA}, nil, nil
	}

	token, err := s.gitHubClient.InstallationToken(ctx, event.Installation.ID)
	if err != nil {
		return review.PullRequestInput{}, "", github.PullRequestMetadata{}, nil, err
	}

	metadata, err := s.gitHubClient.PullRequest(ctx, token, input.Owner, input.Repository, input.Number)
	if err != nil {
		return review.PullRequestInput{}, "", github.PullRequestMetadata{}, nil, err
	}
	files, err := s.gitHubClient.PullRequestFiles(ctx, token, input.Owner, input.Repository, input.Number)
	if err != nil {
		return review.PullRequestInput{}, "", github.PullRequestMetadata{}, nil, err
	}

	input.Title = metadata.Title
	input.Description = metadata.Body
	input.Diff = github.BuildDiff(files)
	if input.Diff == "" {
		input.Diff = event.PullRequest.Head.SHA
	}

	return input, token, metadata, files, nil
}

func (s *Server) publishReview(ctx context.Context, token string, event github.PullRequestEvent, headSHA string, files []github.PullRequestFile, result review.Review) error {
	owner := event.Repository.Owner.Login
	repo := event.Repository.Name
	changedLines := github.ChangedLinesByFile(files)

	for _, finding := range result.Findings {
		if finding.Location.File == "" || finding.Location.Line <= 0 {
			continue
		}
		lines := changedLines[finding.Location.File]
		if _, ok := lines[finding.Location.Line]; !ok {
			continue
		}

		body := formatInlineFinding(finding)
		if err := s.gitHubClient.CreateReviewComment(ctx, token, owner, repo, result.PullRequestNumber, headSHA, finding.Location.File, finding.Location.Line, body); err != nil {
			s.logger.Warn("create inline comment failed", "file", finding.Location.File, "line", finding.Location.Line, "error", err)
		}
	}

	summary := formatReviewSummary(result)
	if err := s.gitHubClient.CreateIssueComment(ctx, token, owner, repo, result.PullRequestNumber, summary); err != nil {
		return err
	}

	conclusion := github.CheckRunNeutral
	if result.OverallRecommendation == "approve" {
		conclusion = github.CheckRunSuccess
	}
	if result.OverallRecommendation == "request_changes" || result.OverallRecommendation == "block_merge" {
		conclusion = github.CheckRunFailure
	}

	return s.gitHubClient.CreateCheckRun(ctx, token, owner, repo, headSHA, conclusion, summary)
}

func (s *Server) getReview(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	s.mu.RLock()
	result, ok := s.reviews[id]
	s.mu.RUnlock()

	if !ok {
		writeError(w, http.StatusNotFound, "review_not_found")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) listRepositories(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	repos := make([]Repository, 0, len(s.repositories))
	for _, repo := range s.repositories {
		repos = append(repos, repo)
	}
	s.mu.RUnlock()

	writeJSON(w, http.StatusOK, repos)
}

func (s *Server) createReviewProfile(w http.ResponseWriter, r *http.Request) {
	var req profile.ReviewProfile
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_payload")
		return
	}

	if strings.TrimSpace(req.ID) == "" || strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "profile_id_and_name_required")
		return
	}

	s.mu.Lock()
	if _, exists := s.profiles[req.ID]; exists {
		s.mu.Unlock()
		writeError(w, http.StatusConflict, "profile_already_exists")
		return
	}
	s.profiles[req.ID] = req
	s.mu.Unlock()

	writeJSON(w, http.StatusCreated, req)
}

func (s *Server) updateReviewProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req profile.ReviewProfile
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_payload")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "profile_name_required")
		return
	}
	req.ID = id

	s.mu.Lock()
	s.profiles[id] = req
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, req)
}

func (s *Server) defaultProfile() profile.ReviewProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.profiles["security-first"]
}

func (s *Server) seenDelivery(deliveryID string) bool {
	if strings.TrimSpace(deliveryID) == "" {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.deliveries[deliveryID]; ok {
		return true
	}
	s.deliveries[deliveryID] = struct{}{}
	return false
}

func decodeJSON(r *http.Request, v any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(v); err != nil {
		return err
	}
	if decoder.Decode(&struct{}{}) == nil {
		return errors.New("multiple JSON values")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]string{"error": code})
}

func formatInlineFinding(finding finding.Finding) string {
	var builder strings.Builder
	builder.WriteString("**")
	builder.WriteString(strings.ToUpper(string(finding.Severity)))
	builder.WriteString(": ")
	builder.WriteString(finding.Title)
	builder.WriteString("**\n\n")
	builder.WriteString(finding.Description)
	if finding.Impact != "" {
		builder.WriteString("\n\nImpact: ")
		builder.WriteString(finding.Impact)
	}
	builder.WriteString("\n\nRecommendation: ")
	builder.WriteString(finding.Recommendation)
	return builder.String()
}

func formatReviewSummary(result review.Review) string {
	return fmt.Sprintf(`# OpenReview Summary

Critical: %d
High: %d
Medium: %d
Low: %d
Informational: %d

Security Score: %d/100

Overall Recommendation: %s
`,
		result.Summary.Critical,
		result.Summary.High,
		result.Summary.Medium,
		result.Summary.Low,
		result.Summary.Informational,
		result.Summary.SecurityScore,
		result.OverallRecommendation,
	)
}
