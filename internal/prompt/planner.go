package prompt

import "github.com/openreview-ai/openreview/internal/profile"

type Job struct {
	PersonaID string
	PromptID  string
}

func Plan(profile profile.ReviewProfile, defaultPersonas []string) []Job {
	personas := profile.Reviewers
	if len(personas) == 0 {
		personas = defaultPersonas
	}

	promptsByPersona := map[string][]string{
		"security-engineer": {
			"security.senior-engineering",
			"security.injection",
			"security.auth",
			"security.authz",
			"security.ssrf",
			"security.xss",
		},
		"staff-backend-engineer": {
			"code.production-readiness",
		},
		"performance-engineer": {
			"code.production-readiness",
		},
	}

	var jobs []Job
	for _, persona := range personas {
		promptIDs := profile.Prompts
		if len(promptIDs) == 0 {
			promptIDs = promptsByPersona[persona]
		}

		for _, promptID := range promptIDs {
			jobs = append(jobs, Job{
				PersonaID: persona,
				PromptID:  promptID,
			})
		}
	}

	return jobs
}
