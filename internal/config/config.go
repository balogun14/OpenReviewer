package config

import (
	"os"
	"strings"
)

type Config struct {
	Addr                string
	GitHubWebhookSecret string
	GitHub              GitHubConfig
	Provider            ProviderConfig
	Retry               RetryConfig
}

type GitHubConfig struct {
	AppID         string
	PrivateKeyPEM []byte
	APIBaseURL    string
}

type ProviderConfig struct {
	Type     string
	BaseURL  string
	APIKey   string
	Model    string
	AppURL   string
	AppTitle string
}

type RetryConfig struct {
	MaxAttempts int
}

func FromEnv() Config {
	loadDotEnv(".env")

	addr := os.Getenv("OPENREVIEW_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	providerType := getenvDefault("OPENREVIEW_PROVIDER", "mock")
	model := os.Getenv("OPENREVIEW_MODEL")
	if model == "" && providerType == "openrouter" {
		model = "anthropic/claude-sonnet-4"
	}

	return Config{
		Addr:                addr,
		GitHubWebhookSecret: os.Getenv("GITHUB_WEBHOOK_SECRET"),
		GitHub: GitHubConfig{
			AppID:         os.Getenv("GITHUB_APP_ID"),
			PrivateKeyPEM: githubPrivateKey(),
			APIBaseURL:    getenvDefault("GITHUB_API_BASE_URL", "https://api.github.com"),
		},
		Provider: ProviderConfig{
			Type:     providerType,
			BaseURL:  os.Getenv("OPENREVIEW_PROVIDER_BASE_URL"),
			APIKey:   providerAPIKey(providerType),
			Model:    model,
			AppURL:   os.Getenv("OPENREVIEW_APP_URL"),
			AppTitle: getenvDefault("OPENREVIEW_APP_TITLE", "OpenReview AI"),
		},
		Retry: RetryConfig{
			MaxAttempts: getenvIntDefault("OPENREVIEW_PROVIDER_MAX_ATTEMPTS", 2),
		},
	}
}

func githubPrivateKey() []byte {
	path := os.Getenv("GITHUB_APP_PRIVATE_KEY_PATH")
	if path != "" {
		content, err := os.ReadFile(path)
		if err == nil {
			return content
		}
	}

	value := os.Getenv("GITHUB_APP_PRIVATE_KEY")
	if value == "" {
		return nil
	}

	normalized := strings.ReplaceAll(value, `\n`, "\n")
	if strings.Contains(normalized, "BEGIN") {
		return []byte(normalized)
	}

	if content, err := os.ReadFile(normalized); err == nil {
		return content
	}

	return []byte(normalized)
}

func providerAPIKey(providerType string) string {
	switch providerType {
	case "openrouter":
		return os.Getenv("OPENROUTER_API_KEY")
	default:
		return os.Getenv("OPENREVIEW_PROVIDER_API_KEY")
	}
}

func getenvDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getenvIntDefault(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	var parsed int
	for _, r := range value {
		if r < '0' || r > '9' {
			return fallback
		}
		parsed = parsed*10 + int(r-'0')
	}
	if parsed <= 0 {
		return fallback
	}
	return parsed
}
