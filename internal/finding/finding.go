package finding

import "strings"

type Severity string

const (
	SeverityCritical      Severity = "critical"
	SeverityHigh          Severity = "high"
	SeverityMedium        Severity = "medium"
	SeverityLow           Severity = "low"
	SeverityInformational Severity = "informational"
)

type Category string

const (
	CategoryArchitecture Category = "architecture"
	CategorySecurity     Category = "security"
	CategoryReliability  Category = "reliability"
	CategoryPerformance  Category = "performance"
	CategoryStyle        Category = "style"
)

type Location struct {
	File string `json:"file,omitempty"`
	Line int    `json:"line,omitempty"`
}

type Finding struct {
	ID             string   `json:"id"`
	Severity       Severity `json:"severity"`
	Category       Category `json:"category"`
	Subcategory    string   `json:"subcategory,omitempty"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Impact         string   `json:"impact,omitempty"`
	Recommendation string   `json:"recommendation"`
	Location       Location `json:"location,omitempty"`
	Confidence     string   `json:"confidence,omitempty"`
	Persona        string   `json:"persona,omitempty"`
}

func Dedupe(findings []Finding) []Finding {
	seen := make(map[string]struct{}, len(findings))
	deduped := make([]Finding, 0, len(findings))

	for _, f := range findings {
		key := strings.ToLower(string(f.Severity) + "|" + string(f.Category) + "|" + f.Title + "|" + f.Location.File)
		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		deduped = append(deduped, f)
	}

	return deduped
}
