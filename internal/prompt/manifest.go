package prompt

type Manifest struct {
	Prompts map[string]Definition
}

type Definition struct {
	ID          string
	Name        string
	Path        string
	Mode        string
	Category    string
	Languages   []string
	Description string
}

func DefaultManifest() Manifest {
	return Manifest{
		Prompts: map[string]Definition{
			"code.production-readiness": {
				ID:          "code.production-readiness",
				Name:        "Production Readiness Review",
				Path:        "curated/code/production-readiness.md",
				Mode:        "review",
				Category:    "quality",
				Languages:   []string{"*"},
				Description: "General production-readiness code review.",
			},
			"security.senior-engineering": {
				ID:          "security.senior-engineering",
				Name:        "Senior Engineering Security Review",
				Path:        "curated/security/senior-engineering.md",
				Mode:        "security",
				Category:    "security",
				Languages:   []string{"*"},
				Description: "Principal-level security, correctness, and reliability review.",
			},
			"security.go": {
				ID:          "security.go",
				Name:        "Go Security Review",
				Path:        "curated/security/go.md",
				Mode:        "security",
				Category:    "security",
				Languages:   []string{"go"},
				Description: "Go-specific security review.",
			},
			"security.injection": {
				ID:          "security.injection",
				Name:        "Injection Review",
				Path:        "curated/security/injection.md",
				Mode:        "security",
				Category:    "injection",
				Languages:   []string{"*"},
				Description: "Defensive source-to-sink injection review.",
			},
			"security.auth": {
				ID:          "security.auth",
				Name:        "Authentication Review",
				Path:        "curated/security/authentication.md",
				Mode:        "security",
				Category:    "authentication",
				Languages:   []string{"*"},
				Description: "Defensive authentication review.",
			},
			"security.authz": {
				ID:          "security.authz",
				Name:        "Authorization Review",
				Path:        "curated/security/authorization.md",
				Mode:        "security",
				Category:    "authorization",
				Languages:   []string{"*"},
				Description: "Defensive authorization and IDOR review.",
			},
			"security.ssrf": {
				ID:          "security.ssrf",
				Name:        "SSRF Review",
				Path:        "curated/security/ssrf.md",
				Mode:        "security",
				Category:    "ssrf",
				Languages:   []string{"*"},
				Description: "Defensive SSRF review.",
			},
			"security.xss": {
				ID:          "security.xss",
				Name:        "XSS Review",
				Path:        "curated/security/xss.md",
				Mode:        "security",
				Category:    "xss",
				Languages:   []string{"*"},
				Description: "Defensive XSS review.",
			},
		},
	}
}
