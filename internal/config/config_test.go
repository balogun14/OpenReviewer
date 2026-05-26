package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGitHubPrivateKeyPrefersPath(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "key.pem")
	if err := os.WriteFile(keyPath, []byte("path-key"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("GITHUB_APP_PRIVATE_KEY_PATH", keyPath)
	t.Setenv("GITHUB_APP_PRIVATE_KEY", "not-raw-pem.pem")

	if got := githubPrivateKey(); string(got.Content) != "path-key" {
		t.Fatalf("expected key from path, got %q", string(got.Content))
	}
}

func TestGitHubPrivateKeyReadsRawPEMWithEscapedNewlines(t *testing.T) {
	t.Setenv("GITHUB_APP_PRIVATE_KEY_PATH", "")
	t.Setenv("GITHUB_APP_PRIVATE_KEY", "-----BEGIN RSA PRIVATE KEY-----\\nabc\\n-----END RSA PRIVATE KEY-----")

	got := string(githubPrivateKey().Content)
	if got != "-----BEGIN RSA PRIVATE KEY-----\nabc\n-----END RSA PRIVATE KEY-----" {
		t.Fatalf("unexpected key content: %q", got)
	}
}

func TestDiagnosticsDoNotExposeSecrets(t *testing.T) {
	cfg := Config{
		Addr:                ":8080",
		GitHubWebhookSecret: "secret",
		GitHub: GitHubConfig{
			AppID:            "123",
			PrivateKeyPEM:    []byte("private-key"),
			PrivateKeySource: "GITHUB_APP_PRIVATE_KEY_PATH",
			PrivateKeyPath:   "key.pem",
			PrivateKeyExists: true,
			APIBaseURL:       "https://api.github.com",
		},
		Provider: ProviderConfig{
			Type:   "openrouter",
			APIKey: "provider-secret",
			Model:  "model",
		},
		Retry: RetryConfig{MaxAttempts: 2},
	}

	diagnostics := cfg.Diagnostics()
	if diagnostics["github_webhook_secret_set"] != true {
		t.Fatal("expected webhook secret boolean")
	}
	if diagnostics["provider_api_key_set"] != true {
		t.Fatal("expected provider key boolean")
	}
	for _, forbidden := range []string{"secret", "provider-secret", "private-key"} {
		for key, value := range diagnostics {
			if value == forbidden {
				t.Fatalf("diagnostic %s exposed secret value", key)
			}
		}
	}
}
