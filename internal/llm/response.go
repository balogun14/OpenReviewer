package llm

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/openreview-ai/openreview/internal/finding"
)

type ReviewResponse struct {
	Findings       []FindingResponse `json:"findings"`
	Summary        string            `json:"summary"`
	Recommendation string            `json:"recommendation"`
}

type FindingResponse struct {
	Severity       string `json:"severity"`
	Category       string `json:"category"`
	Subcategory    string `json:"subcategory"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Impact         string `json:"impact"`
	Recommendation string `json:"recommendation"`
	File           string `json:"file"`
	Line           int    `json:"line"`
	Confidence     string `json:"confidence"`
}

type ResponseContext struct {
	PromptID  string
	PersonaID string
}

type ParsedReview struct {
	Findings       []finding.Finding
	Summary        string
	Recommendation string
}

type ValidationError struct {
	Problems []string
}

func (e ValidationError) Error() string {
	return "invalid LLM review response: " + strings.Join(e.Problems, "; ")
}

func ParseReviewResponse(raw string, context ResponseContext) (ParsedReview, error) {
	payload, err := extractJSONObject(raw)
	if err != nil {
		return ParsedReview{}, err
	}

	var response ReviewResponse
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&response); err != nil {
		return ParsedReview{}, fmt.Errorf("decode LLM review response: %w", err)
	}

	parsed := ParsedReview{
		Summary:        strings.TrimSpace(response.Summary),
		Recommendation: normalizeRecommendation(response.Recommendation),
	}

	var problems []string
	for i, rawFinding := range response.Findings {
		normalized, findingProblems := normalizeFinding(rawFinding, context)
		for _, problem := range findingProblems {
			problems = append(problems, fmt.Sprintf("findings[%d].%s", i, problem))
		}
		if len(findingProblems) == 0 {
			parsed.Findings = append(parsed.Findings, normalized)
		}
	}

	if parsed.Recommendation == "" {
		problems = append(problems, "recommendation must be one of approve, comment, request_changes, block_merge")
	}

	if len(problems) > 0 {
		return ParsedReview{}, ValidationError{Problems: problems}
	}

	return parsed, nil
}

func extractJSONObject(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", errors.New("empty LLM review response")
	}

	if strings.HasPrefix(value, "```") {
		lines := strings.Split(value, "\n")
		if len(lines) >= 3 {
			lines = lines[1 : len(lines)-1]
			value = strings.TrimSpace(strings.Join(lines, "\n"))
		}
	}

	start := strings.Index(value, "{")
	end := strings.LastIndex(value, "}")
	if start < 0 || end < start {
		return "", errors.New("LLM review response does not contain a JSON object")
	}

	return value[start : end+1], nil
}

func normalizeFinding(raw FindingResponse, context ResponseContext) (finding.Finding, []string) {
	var problems []string

	severity := normalizeSeverity(raw.Severity)
	if severity == "" {
		problems = append(problems, "severity must be one of critical, high, medium, low, informational")
	}

	category := normalizeCategory(raw.Category)
	if category == "" {
		problems = append(problems, "category must be one of security, architecture, reliability, performance, style")
	}

	title := strings.TrimSpace(raw.Title)
	if title == "" {
		problems = append(problems, "title is required")
	}

	description := strings.TrimSpace(raw.Description)
	if description == "" {
		problems = append(problems, "description is required")
	}

	recommendation := strings.TrimSpace(raw.Recommendation)
	if recommendation == "" {
		problems = append(problems, "recommendation is required")
	}

	line := raw.Line
	if line < 0 {
		problems = append(problems, "line cannot be negative")
	}

	confidence := normalizeConfidence(raw.Confidence)
	if confidence == "" && strings.TrimSpace(raw.Confidence) != "" {
		problems = append(problems, "confidence must be one of high, medium, low")
	}

	normalized := finding.Finding{
		ID:             findingID(context, raw),
		Severity:       severity,
		Category:       category,
		Subcategory:    strings.TrimSpace(raw.Subcategory),
		Title:          title,
		Description:    description,
		Impact:         strings.TrimSpace(raw.Impact),
		Recommendation: recommendation,
		Location: finding.Location{
			File: strings.TrimSpace(raw.File),
			Line: line,
		},
		Confidence: confidence,
		Persona:    context.PersonaID,
	}

	return normalized, problems
}

func normalizeSeverity(value string) finding.Severity {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "critical", "blocker":
		return finding.SeverityCritical
	case "high":
		return finding.SeverityHigh
	case "medium", "important":
		return finding.SeverityMedium
	case "low", "minor":
		return finding.SeverityLow
	case "informational", "info":
		return finding.SeverityInformational
	default:
		return ""
	}
}

func normalizeCategory(value string) finding.Category {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "security":
		return finding.CategorySecurity
	case "architecture":
		return finding.CategoryArchitecture
	case "reliability", "correctness", "testing":
		return finding.CategoryReliability
	case "performance":
		return finding.CategoryPerformance
	case "style", "maintainability":
		return finding.CategoryStyle
	default:
		return ""
	}
}

func normalizeRecommendation(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "approve":
		return "approve"
	case "comment", "approve_with_comments", "approve with comments":
		return "comment"
	case "request_changes", "request changes":
		return "request_changes"
	case "block_merge", "block merge", "block":
		return "block_merge"
	default:
		return ""
	}
}

func normalizeConfidence(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "":
		return ""
	case "high":
		return "high"
	case "medium", "med":
		return "medium"
	case "low":
		return "low"
	default:
		return ""
	}
}

func findingID(context ResponseContext, raw FindingResponse) string {
	key := strings.Join([]string{
		context.PromptID,
		context.PersonaID,
		raw.Severity,
		raw.Category,
		raw.Title,
		raw.File,
		fmt.Sprintf("%d", raw.Line),
	}, "|")
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:6])
}
