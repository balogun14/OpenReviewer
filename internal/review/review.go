package review

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/openreview-ai/openreview/internal/finding"
	"github.com/openreview-ai/openreview/internal/profile"
	"github.com/openreview-ai/openreview/internal/prompt"
	"github.com/openreview-ai/openreview/internal/provider"
)

type PullRequestInput struct {
	Owner       string                `json:"owner"`
	Repository  string                `json:"repository"`
	Number      int                   `json:"number"`
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Diff        string                `json:"diff"`
	Profile     profile.ReviewProfile `json:"profile"`
}

type ReviewerPersona struct {
	ID           string   `json:"id"`
	DisplayName  string   `json:"displayName"`
	Instructions []string `json:"instructions"`
}

type Review struct {
	ID                    string            `json:"id"`
	Repository            string            `json:"repository"`
	PullRequestNumber     int               `json:"pullRequestNumber"`
	Summary               Summary           `json:"summary"`
	Findings              []finding.Finding `json:"findings"`
	OverallRecommendation string            `json:"overallRecommendation"`
	CreatedAt             time.Time         `json:"createdAt"`
}

type Summary struct {
	Critical      int `json:"critical"`
	High          int `json:"high"`
	Medium        int `json:"medium"`
	Low           int `json:"low"`
	Informational int `json:"informational"`
	SecurityScore int `json:"securityScore"`
}

type Provider interface {
	Review(ctx context.Context, req provider.Request) ([]finding.Finding, error)
}

type Engine struct {
	provider       Provider
	promptRenderer prompt.Renderer
	personas       map[string]ReviewerPersona
}

func NewEngine(provider Provider, promptRenderer prompt.Renderer, personas []ReviewerPersona) *Engine {
	byID := make(map[string]ReviewerPersona, len(personas))
	for _, persona := range personas {
		byID[persona.ID] = persona
	}

	return &Engine{
		provider:       provider,
		promptRenderer: promptRenderer,
		personas:       byID,
	}
}

func (e *Engine) ReviewPullRequest(ctx context.Context, input PullRequestInput) (Review, error) {
	if input.Profile.ID == "" {
		input.Profile = profile.SecurityFirst()
	}

	var allFindings []finding.Finding
	for _, job := range prompt.Plan(input.Profile, []string{"security-engineer", "staff-backend-engineer"}) {
		persona, ok := e.personas[job.PersonaID]
		if !ok {
			continue
		}

		rendered, err := e.promptRenderer.Render(job.PromptID, promptVariables(input, persona))
		if err != nil {
			return Review{}, fmt.Errorf("render prompt %q for persona %q: %w", job.PromptID, job.PersonaID, err)
		}

		findings, err := e.provider.Review(ctx, provider.Request{
			Owner:               input.Owner,
			Repository:          input.Repository,
			PullRequestNumber:   input.Number,
			PullRequestTitle:    input.Title,
			PullRequestBody:     input.Description,
			Diff:                input.Diff,
			ProfileRules:        input.Profile.Rules,
			PersonaID:           persona.ID,
			PersonaDisplayName:  persona.DisplayName,
			PersonaInstructions: persona.Instructions,
			PromptID:            rendered.Definition.ID,
			PromptName:          rendered.Definition.Name,
			Prompt:              rendered.Content,
		})
		if err != nil {
			return Review{}, fmt.Errorf("review with persona %q prompt %q: %w", job.PersonaID, job.PromptID, err)
		}

		for i := range findings {
			findings[i].Persona = persona.ID
		}
		allFindings = append(allFindings, findings...)
	}

	allFindings = finding.Dedupe(allFindings)
	summary := summarize(allFindings)

	return Review{
		ID:                    reviewID(input),
		Repository:            input.Owner + "/" + input.Repository,
		PullRequestNumber:     input.Number,
		Summary:               summary,
		Findings:              allFindings,
		OverallRecommendation: recommendation(summary),
		CreatedAt:             time.Now().UTC(),
	}, nil
}

func promptVariables(input PullRequestInput, persona ReviewerPersona) prompt.Variables {
	return prompt.Variables{
		"REPO_NAME":            input.Owner + "/" + input.Repository,
		"PR_NUMBER":            fmt.Sprintf("%d", input.Number),
		"PR_TITLE":             input.Title,
		"PR_DESCRIPTION":       input.Description,
		"DIFF":                 input.Diff,
		"DESCRIPTION":          input.Description,
		"WHAT_WAS_IMPLEMENTED": input.Title,
		"PLAN_OR_REQUIREMENTS": stringsFromList(input.Profile.Rules),
		"PLAN_REFERENCE":       stringsFromList(input.Profile.Rules),
		"BASE_SHA":             "unknown",
		"HEAD_SHA":             "unknown",
		"REVIEW_PROFILE_RULES": stringsFromList(input.Profile.Rules),
		"PERSONA_ID":           persona.ID,
		"PERSONA_NAME":         persona.DisplayName,
		"PERSONA_INSTRUCTIONS": stringsFromList(persona.Instructions),
		"LOGIN_INSTRUCTIONS":   "Not applicable. This is a defensive pull request code review, not a live application test.",
		"RULES_AVOID":          "Do not perform live exploitation or destructive testing.",
		"RULES_FOCUS":          stringsFromList(input.Profile.Rules),
		"WEB_URL":              "Not applicable.",
		"REPO_PATH":            input.Owner + "/" + input.Repository,
		"PLAYWRIGHT_SESSION":   "not-applicable",
		"AUTH_CONTEXT":         "Not applicable. Review code and diff only.",
		"CHANGED_FILES":        "Diff-only context provided.",
		"REPOSITORY_RULES":     stringsFromList(input.Profile.Rules),
		"LANGUAGE_CONTEXT":     "Infer language and framework from changed files.",
		"OUTPUT_SCHEMA":        "Use the OpenReview JSON output contract below.",
	}
}

func stringsFromList(values []string) string {
	if len(values) == 0 {
		return "None provided."
	}

	return "- " + strings.Join(values, "\n- ")
}

func DefaultReviewerPersonas() []ReviewerPersona {
	return []ReviewerPersona{
		{
			ID:          "security-engineer",
			DisplayName: "Security Engineer",
			Instructions: []string{
				"prioritize authentication and authorization risks",
				"flag injection, SSRF, XSS, CSRF, secrets, cryptography, and unsafe file upload risks",
			},
		},
		{
			ID:          "staff-backend-engineer",
			DisplayName: "Staff Backend Engineer",
			Instructions: []string{
				"evaluate correctness, maintainability, error handling, and architectural fit",
				"avoid low-value style comments unless they affect long-term maintainability",
			},
		},
		{
			ID:          "performance-engineer",
			DisplayName: "Performance Engineer",
			Instructions: []string{
				"flag avoidable N+1 queries, unbounded memory use, inefficient algorithms, and missing timeouts",
			},
		},
	}
}

func summarize(findings []finding.Finding) Summary {
	summary := Summary{SecurityScore: 100}

	for _, f := range findings {
		switch f.Severity {
		case finding.SeverityCritical:
			summary.Critical++
		case finding.SeverityHigh:
			summary.High++
		case finding.SeverityMedium:
			summary.Medium++
		case finding.SeverityLow:
			summary.Low++
		case finding.SeverityInformational:
			summary.Informational++
		}

		if f.Category == finding.CategorySecurity {
			switch f.Severity {
			case finding.SeverityCritical:
				summary.SecurityScore -= 30
			case finding.SeverityHigh:
				summary.SecurityScore -= 20
			case finding.SeverityMedium:
				summary.SecurityScore -= 10
			case finding.SeverityLow:
				summary.SecurityScore -= 5
			}
		}
	}

	if summary.SecurityScore < 0 {
		summary.SecurityScore = 0
	}

	return summary
}

func recommendation(summary Summary) string {
	if summary.Critical > 0 || summary.High > 0 {
		return "request_changes"
	}
	if summary.Medium > 0 {
		return "comment"
	}
	return "approve"
}

func reviewID(input PullRequestInput) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s/%s#%d:%s", input.Owner, input.Repository, input.Number, input.Diff)))
	return hex.EncodeToString(sum[:8])
}
