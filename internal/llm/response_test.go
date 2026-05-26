package llm

import (
	"errors"
	"testing"

	"github.com/openreview-ai/openreview/internal/finding"
)

func TestParseReviewResponseNormalizesFindings(t *testing.T) {
	raw := `{
		"findings": [
			{
				"severity": "High",
				"category": "Correctness",
				"subcategory": "timeouts",
				"title": "Missing outbound request timeout",
				"description": "The HTTP client can wait forever.",
				"impact": "A slow dependency can exhaust worker capacity.",
				"recommendation": "Use a client timeout or request context deadline.",
				"file": "internal/client.go",
				"line": 42,
				"confidence": "med"
			}
		],
		"summary": "One issue found.",
		"recommendation": "request changes"
	}`

	parsed, err := ParseReviewResponse(raw, ResponseContext{
		PromptID:  "security.go",
		PersonaID: "security-engineer",
	})
	if err != nil {
		t.Fatalf("ParseReviewResponse returned error: %v", err)
	}

	if parsed.Recommendation != "request_changes" {
		t.Fatalf("unexpected recommendation: %q", parsed.Recommendation)
	}
	if len(parsed.Findings) != 1 {
		t.Fatalf("expected one finding, got %d", len(parsed.Findings))
	}

	got := parsed.Findings[0]
	if got.ID == "" {
		t.Fatal("expected stable finding ID")
	}
	if got.Severity != finding.SeverityHigh {
		t.Fatalf("unexpected severity: %q", got.Severity)
	}
	if got.Category != finding.CategoryReliability {
		t.Fatalf("unexpected category: %q", got.Category)
	}
	if got.Confidence != "medium" {
		t.Fatalf("unexpected confidence: %q", got.Confidence)
	}
	if got.Persona != "security-engineer" {
		t.Fatalf("unexpected persona: %q", got.Persona)
	}
}

func TestParseReviewResponseAcceptsFencedJSON(t *testing.T) {
	raw := "```json\n{\"findings\":[],\"summary\":\"none\",\"recommendation\":\"approve\"}\n```"

	parsed, err := ParseReviewResponse(raw, ResponseContext{})
	if err != nil {
		t.Fatalf("ParseReviewResponse returned error: %v", err)
	}
	if parsed.Recommendation != "approve" {
		t.Fatalf("unexpected recommendation: %q", parsed.Recommendation)
	}
}

func TestParseReviewResponseRejectsMalformedJSON(t *testing.T) {
	_, err := ParseReviewResponse("not json", ResponseContext{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseReviewResponseRejectsInvalidFinding(t *testing.T) {
	raw := `{
		"findings": [
			{
				"severity": "urgent",
				"category": "security",
				"title": "",
				"description": "desc",
				"recommendation": "fix",
				"line": -1
			}
		],
		"summary": "bad",
		"recommendation": "approve"
	}`

	_, err := ParseReviewResponse(raw, ResponseContext{})
	if err == nil {
		t.Fatal("expected validation error")
	}

	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if len(validationErr.Problems) != 3 {
		t.Fatalf("expected three validation problems, got %d: %v", len(validationErr.Problems), validationErr.Problems)
	}
}
