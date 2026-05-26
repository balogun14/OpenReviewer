package prompt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/openreview-ai/openreview/internal/profile"
)

func TestRendererInterpolatesVariablesAndAddsContract(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "review.txt"), []byte("Review {{REPO_NAME}} {PR_TITLE}"), 0o644); err != nil {
		t.Fatal(err)
	}

	renderer := NewRenderer(NewLoader(dir), Manifest{
		Prompts: map[string]Definition{
			"test.review": {
				ID:       "test.review",
				Name:     "Test Review",
				Path:     "review.txt",
				Category: "quality",
			},
		},
	})

	rendered, err := renderer.Render("test.review", Variables{
		"REPO_NAME": "acme/api",
		"PR_TITLE":  "Add auth",
	})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	if rendered.Definition.ID != "test.review" {
		t.Fatalf("unexpected definition: %q", rendered.Definition.ID)
	}
	if contains := "Review acme/api Add auth"; !strings.Contains(rendered.Content, contains) {
		t.Fatalf("expected rendered content to contain %q:\n%s", contains, rendered.Content)
	}
	if !strings.Contains(rendered.Content, "openreview_output_contract") {
		t.Fatalf("expected output contract:\n%s", rendered.Content)
	}
}

func TestPlanUsesPersonaDefaults(t *testing.T) {
	jobs := Plan(profile.ReviewProfile{}, []string{"security-engineer"})
	if len(jobs) != 6 {
		t.Fatalf("expected six security jobs, got %d", len(jobs))
	}
	if jobs[0].PersonaID != "security-engineer" {
		t.Fatalf("unexpected persona: %q", jobs[0].PersonaID)
	}
}
