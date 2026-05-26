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

	if got := string(githubPrivateKey()); got != "path-key" {
		t.Fatalf("expected key from path, got %q", got)
	}
}

func TestGitHubPrivateKeyReadsRawPEMWithEscapedNewlines(t *testing.T) {
	t.Setenv("GITHUB_APP_PRIVATE_KEY_PATH", "")
	t.Setenv("GITHUB_APP_PRIVATE_KEY", "-----BEGIN RSA PRIVATE KEY-----\\nabc\\n-----END RSA PRIVATE KEY-----")

	got := string(githubPrivateKey())
	if got != "-----BEGIN RSA PRIVATE KEY-----\nabc\n-----END RSA PRIVATE KEY-----" {
		t.Fatalf("unexpected key content: %q", got)
	}
}
