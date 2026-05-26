package review

import (
	"context"
	"testing"

	"github.com/openreview-ai/openreview/internal/prompt"
	"github.com/openreview-ai/openreview/internal/provider"
)

func TestReviewPullRequestRunsPersonasAndSummarizes(t *testing.T) {
	renderer := prompt.NewRenderer(prompt.NewLoader("../../prompts"), prompt.DefaultManifest())
	engine := NewEngine(provider.NewMockProvider(), renderer, DefaultReviewerPersonas())

	result, err := engine.ReviewPullRequest(context.Background(), PullRequestInput{
		Owner:      "acme",
		Repository: "api",
		Number:     42,
		Diff:       "+ password := \"secret\"\n+ // TODO: replace this",
	})
	if err != nil {
		t.Fatalf("ReviewPullRequest returned error: %v", err)
	}

	if result.ID == "" {
		t.Fatal("expected review ID")
	}
	if result.Summary.High != 1 {
		t.Fatalf("expected one high finding after dedupe, got %d", result.Summary.High)
	}
	if result.Summary.Informational != 1 {
		t.Fatalf("expected one informational finding after dedupe, got %d", result.Summary.Informational)
	}
	if result.Summary.SecurityScore != 80 {
		t.Fatalf("expected security score 80, got %d", result.Summary.SecurityScore)
	}
	if result.OverallRecommendation != "request_changes" {
		t.Fatalf("expected request_changes, got %q", result.OverallRecommendation)
	}
}
