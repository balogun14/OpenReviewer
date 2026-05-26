package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/openreview-ai/openreview/internal/finding"
)

type MockProvider struct{}

func NewMockProvider() MockProvider {
	return MockProvider{}
}

func (p MockProvider) Review(ctx context.Context, req Request) ([]finding.Finding, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	diff := strings.ToLower(req.Diff)
	var findings []finding.Finding

	if strings.Contains(diff, "password") || strings.Contains(diff, "api_key") || strings.Contains(diff, "secret") {
		findings = append(findings, newFinding(
			finding.SeverityHigh,
			finding.CategorySecurity,
			"Possible secret exposure",
			"The change appears to introduce sensitive credential material.",
			"Move secrets to a managed secret store and rotate any exposed values.",
		))
	}

	if strings.Contains(diff, "select ") && strings.Contains(diff, "+") {
		findings = append(findings, newFinding(
			finding.SeverityHigh,
			finding.CategorySecurity,
			"Possible SQL injection",
			"The change appears to build SQL with string concatenation.",
			"Use parameterized queries or a query builder that binds user input safely.",
		))
	}

	if strings.Contains(diff, "todo") {
		findings = append(findings, newFinding(
			finding.SeverityInformational,
			finding.CategoryReliability,
			"TODO left in changed code",
			"The change includes a TODO that may represent incomplete review-critical work.",
			"Resolve the TODO or link it to a tracked follow-up if it is intentionally deferred.",
		))
	}

	return findings, nil
}

func newFinding(severity finding.Severity, category finding.Category, title, description, recommendation string) finding.Finding {
	sum := sha256.Sum256([]byte(string(severity) + "|" + string(category) + "|" + title))
	return finding.Finding{
		ID:             hex.EncodeToString(sum[:6]),
		Severity:       severity,
		Category:       category,
		Title:          title,
		Description:    description,
		Recommendation: recommendation,
	}
}
