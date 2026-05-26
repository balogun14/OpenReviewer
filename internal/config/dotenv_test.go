package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnvDoesNotOverrideExistingEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("OPENREVIEW_ADDR=:9999\nEXISTING=from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("EXISTING", "from-env")
	loadDotEnv(path)

	if got := os.Getenv("OPENREVIEW_ADDR"); got != ":9999" {
		t.Fatalf("expected env file value, got %q", got)
	}
	if got := os.Getenv("EXISTING"); got != "from-env" {
		t.Fatalf("expected existing env to win, got %q", got)
	}
}
