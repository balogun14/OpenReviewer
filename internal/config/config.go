package config

import (
	"os"
	"strings"
)

type Config struct {
	Addr                string
	DotEnv              DotEnvStatus
	GitHubWebhookSecret string
	GitHub              GitHubConfig
	Provider            ProviderConfig
	Retry               RetryConfig
}

type GitHubConfig struct {
	AppID            string
	PrivateKeyPEM    []byte
	PrivateKeySource string
	PrivateKeyPath   string
	PrivateKeyExists bool
	APIBaseURL       string
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
	dotEnv := loadDotEnv(".env")

	addr := os.Getenv("OPENREVIEW_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	providerType := getenvDefault("OPENREVIEW_PROVIDER", "mock")
	model := os.Getenv("OPENREVIEW_MODEL")
	if model == "" && providerType == "openrouter" {
		model = "anthropic/claude-sonnet-4"
	}

	privateKey := githubPrivateKey()

	return Config{
		Addr:                addr,
		DotEnv:              dotEnv,
		GitHubWebhookSecret: os.Getenv("GITHUB_WEBHOOK_SECRET"),
		GitHub: GitHubConfig{
			AppID:            os.Getenv("GITHUB_APP_ID"),
			PrivateKeyPEM:    privateKey.Content,
			PrivateKeySource: privateKey.Source,
			PrivateKeyPath:   privateKey.Path,
			PrivateKeyExists: privateKey.Exists,
			APIBaseURL:       getenvDefault("GITHUB_API_BASE_URL", "https://api.github.com"),
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

type privateKeyResult struct {
	Content []byte
	Source  string
	Path    string
	Exists  bool
}

func githubPrivateKey() privateKeyResult {
	path := os.Getenv("GITHUB_APP_PRIVATE_KEY_PATH")
	if path != "" {
		content, err := os.ReadFile(path)
		if err == nil {
			return privateKeyResult{Content: content, Source: "GITHUB_APP_PRIVATE_KEY_PATH", Path: path, Exists: true}
		}
		return privateKeyResult{Source: "GITHUB_APP_PRIVATE_KEY_PATH", Path: path, Exists: false}
	}

	value := os.Getenv("GITHUB_APP_PRIVATE_KEY")
	if value == "" {
		return privateKeyResult{}
	}

	normalized := strings.ReplaceAll(value, `\n`, "\n")
	if strings.Contains(normalized, "BEGIN") {
		return privateKeyResult{Content: []byte(normalized), Source: "GITHUB_APP_PRIVATE_KEY", Exists: true}
	}

	if content, err := os.ReadFile(normalized); err == nil {
		return privateKeyResult{Content: content, Source: "GITHUB_APP_PRIVATE_KEY", Path: normalized, Exists: true}
	}

	return privateKeyResult{Content: []byte(normalized), Source: "GITHUB_APP_PRIVATE_KEY", Path: normalized, Exists: false}
}

func (c Config) Diagnostics() map[string]any {
	return map[string]any{
		"dotenv_path":                    c.DotEnv.Path,
		"dotenv_loaded":                  c.DotEnv.Loaded,
		"addr":                           c.Addr,
		"github_webhook_secret_set":      c.GitHubWebhookSecret != "",
		"github_app_id_set":              c.GitHub.AppID != "",
		"github_private_key_source":      emptyAsUnset(c.GitHub.PrivateKeySource),
		"github_private_key_path":        emptyAsUnset(c.GitHub.PrivateKeyPath),
		"github_private_key_file_exists": c.GitHub.PrivateKeyExists,
		"github_private_key_loaded":      len(c.GitHub.PrivateKeyPEM) > 0,
		"github_api_base_url":            c.GitHub.APIBaseURL,
		"provider_type":                  c.Provider.Type,
		"provider_model":                 emptyAsUnset(c.Provider.Model),
		"provider_base_url":              emptyAsUnset(c.Provider.BaseURL),
		"provider_api_key_set":           c.Provider.APIKey != "",
		"provider_max_attempts":          c.Retry.MaxAttempts,
	}
}

func emptyAsUnset(value string) string {
	if value == "" {
		return "unset"
	}
	return value
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
