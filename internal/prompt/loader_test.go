package prompt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoaderExpandsIncludes(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "shared"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "shared", "rules.txt"), []byte("shared rules"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.txt"), []byte("before @include(shared/rules.txt) after"), 0o644); err != nil {
		t.Fatal(err)
	}

	content, err := NewLoader(dir).Load(Definition{ID: "test", Path: "main.txt"})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if content != "before shared rules after" {
		t.Fatalf("unexpected content: %q", content)
	}
}

func TestLoaderRejectsPathTraversal(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.txt"), []byte("@include(../outside.txt)"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := NewLoader(dir).Load(Definition{ID: "test", Path: "main.txt"})
	if err == nil {
		t.Fatal("expected path traversal error")
	}
	if !strings.Contains(err.Error(), "escapes base directory") {
		t.Fatalf("unexpected error: %v", err)
	}
}
