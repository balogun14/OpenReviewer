package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

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
	reviewProvider := buildProvider(cfg, logger)
	engine := review.NewEngine(reviewProvider, promptRenderer, review.DefaultReviewerPersonas())
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

func buildProvider(cfg config.Config, logger *slog.Logger) provider.Reviewer {
	var reviewProvider provider.Reviewer

	switch cfg.Provider.Type {
	case "openrouter":
		reviewProvider = provider.NewOpenRouterProvider(
			cfg.Provider.APIKey,
			cfg.Provider.Model,
			cfg.Provider.AppURL,
			cfg.Provider.AppTitle,
		)
	case "openai-compatible":
		reviewProvider = provider.NewOpenAICompatibleProvider(provider.OpenAICompatibleConfig{
			BaseURL: cfg.Provider.BaseURL,
			APIKey:  cfg.Provider.APIKey,
			Model:   cfg.Provider.Model,
		})
	default:
		logger.Warn("using mock provider", "provider", cfg.Provider.Type)
		reviewProvider = provider.NewMockProvider()
	}

	return provider.WithRetry(reviewProvider, provider.RetryConfig{
		MaxAttempts: cfg.Retry.MaxAttempts,
		Delay:       500 * time.Millisecond,
	})
}
