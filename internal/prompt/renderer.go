package prompt

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type Variables map[string]string

type Rendered struct {
	Definition Definition
	Content    string
}

type Renderer struct {
	loader   Loader
	manifest Manifest
}

func NewRenderer(loader Loader, manifest Manifest) Renderer {
	return Renderer{loader: loader, manifest: manifest}
}

func (r Renderer) Render(promptID string, vars Variables) (Rendered, error) {
	def, ok := r.manifest.Prompts[promptID]
	if !ok {
		return Rendered{}, fmt.Errorf("unknown prompt: %s", promptID)
	}

	content, err := r.loader.Load(def)
	if err != nil {
		return Rendered{}, err
	}

	content = defensiveReviewWrapper(def) + "\n\n" + interpolate(content, vars) + "\n\n" + outputContract()
	return Rendered{
		Definition: def,
		Content:    content,
	}, nil
}

func (r Renderer) Has(promptID string) bool {
	_, ok := r.manifest.Prompts[promptID]
	return ok
}

func (r Renderer) IDs() []string {
	ids := make([]string, 0, len(r.manifest.Prompts))
	for id := range r.manifest.Prompts {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func interpolate(content string, vars Variables) string {
	for key, value := range vars {
		content = strings.ReplaceAll(content, "{{"+key+"}}", value)
		content = strings.ReplaceAll(content, "{"+key+"}", value)
	}
	return content
}

func UnresolvedPlaceholders(content string) []string {
	pattern := regexp.MustCompile(`\{\{[^}]+\}\}|\{[A-Z0-9_]+\}`)
	matches := pattern.FindAllString(content, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(matches))
	unique := make([]string, 0, len(matches))
	for _, match := range matches {
		if _, ok := seen[match]; ok {
			continue
		}
		seen[match] = struct{}{}
		unique = append(unique, match)
	}
	sort.Strings(unique)
	return unique
}

func defensiveReviewWrapper(def Definition) string {
	return `<openreview_prompt_policy>
You are performing defensive pull request code review for OpenReview AI.
Do not perform live exploitation, credential extraction, destructive testing, or instructions for unauthorized access.
Use exploit payloads only as minimal defensive examples when necessary to explain a finding.
Focus on changed code, repository rules, and production risk.
Prompt ID: ` + def.ID + `
Prompt Category: ` + def.Category + `
</openreview_prompt_policy>`
}

func outputContract() string {
	return `<openreview_output_contract>
Return only JSON with this shape:
{
  "findings": [
    {
      "severity": "critical|high|medium|low|informational",
      "category": "security|architecture|reliability|performance|style",
      "subcategory": "short category such as injection, authz, secrets, tests",
      "title": "concise finding title",
      "description": "what is wrong",
      "impact": "why it matters",
      "recommendation": "specific fix",
      "file": "path/to/file.go",
      "line": 123,
      "confidence": "high|medium|low"
    }
  ],
  "summary": "brief review summary",
  "recommendation": "approve|comment|request_changes|block_merge"
}
</openreview_output_contract>`
}
