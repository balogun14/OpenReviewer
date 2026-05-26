package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/openreview-ai/openreview/internal/config"
	"github.com/openreview-ai/openreview/internal/github"
	"github.com/openreview-ai/openreview/internal/httpapi"
	"github.com/openreview-ai/openreview/internal/prompt"
	"github.com/openreview-ai/openreview/internal/provider"
	"github.com/openreview-ai/openreview/internal/review"
)

func main() {
	cfg := config.FromEnv()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("configuration loaded", "diagnostics", cfg.Diagnostics())

	promptRenderer := prompt.NewRenderer(prompt.NewLoader("prompts"), prompt.DefaultManifest())
	reviewProvider := buildProvider(cfg, logger)
	engine := review.NewEngine(reviewProvider, promptRenderer, review.DefaultReviewerPersonas())
	gitHubClient := buildGitHubClient(cfg)
	if gitHubClient == nil {
		logger.Warn("GitHub App integration disabled; set GITHUB_APP_ID and GITHUB_APP_PRIVATE_KEY_PATH")
	} else {
		logger.Info("GitHub App integration enabled", "apiBaseUrl", cfg.GitHub.APIBaseURL)
	}
	server := httpapi.NewServer(httpapi.ServerOptions{
		Logger:              logger,
		ReviewEngine:        engine,
		GitHubWebhookSecret: cfg.GitHubWebhookSecret,
		GitHubClient:        gitHubClient,
	})

	logger.Info("starting OpenReview AI", "addr", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, server.Routes()); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func buildGitHubClient(cfg config.Config) *github.Client {
	if cfg.GitHub.AppID == "" || len(cfg.GitHub.PrivateKeyPEM) == 0 {
		return nil
	}

	return github.NewClient(github.ClientConfig{
		BaseURL:       cfg.GitHub.APIBaseURL,
		AppID:         cfg.GitHub.AppID,
		PrivateKeyPEM: cfg.GitHub.PrivateKeyPEM,
	})
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
