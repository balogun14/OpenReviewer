package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/openreview-ai/openreview/internal/config"
	"github.com/openreview-ai/openreview/internal/httpapi"
	"github.com/openreview-ai/openreview/internal/prompt"
	"github.com/openreview-ai/openreview/internal/provider"
	"github.com/openreview-ai/openreview/internal/review"
)

func main() {
	cfg := config.FromEnv()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	promptRenderer := prompt.NewRenderer(prompt.NewLoader("prompts"), prompt.DefaultManifest())
	engine := review.NewEngine(provider.NewMockProvider(), promptRenderer, review.DefaultReviewerPersonas())
	server := httpapi.NewServer(httpapi.ServerOptions{
		Logger:              logger,
		ReviewEngine:        engine,
		GitHubWebhookSecret: cfg.GitHubWebhookSecret,
	})

	logger.Info("starting OpenReview AI", "addr", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, server.Routes()); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
