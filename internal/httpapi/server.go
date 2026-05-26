package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/openreview-ai/openreview/internal/github"
	"github.com/openreview-ai/openreview/internal/profile"
	"github.com/openreview-ai/openreview/internal/review"
)

type ServerOptions struct {
	Logger              *slog.Logger
	ReviewEngine        *review.Engine
	GitHubWebhookSecret string
}

type Server struct {
	logger              *slog.Logger
	reviewEngine        *review.Engine
	gitHubWebhookSecret string
	mu                  sync.RWMutex
	reviews             map[string]review.Review
	profiles            map[string]profile.ReviewProfile
	repositories        map[string]Repository
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
		reviews:             make(map[string]review.Review),
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

	input := review.PullRequestInput{
		Owner:       event.Repository.Owner.Login,
		Repository:  event.Repository.Name,
		Number:      event.Number,
		Title:       event.PullRequest.Title,
		Description: event.PullRequest.Body,
		Diff:        event.PullRequest.Head.SHA,
		Profile:     s.defaultProfile(),
	}

	result, err := s.reviewEngine.ReviewPullRequest(r.Context(), input)
	if err != nil {
		s.logger.Error("review failed", "error", err)
		writeError(w, http.StatusInternalServerError, "review_failed")
		return
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
