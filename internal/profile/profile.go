package profile

type ReviewProfile struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Rules     []string `json:"rules"`
	Reviewers []string `json:"reviewers"`
	Prompts   []string `json:"prompts,omitempty"`
}

func SecurityFirst() ReviewProfile {
	return ReviewProfile{
		ID:   "security-first",
		Name: "Security First",
		Rules: []string{
			"prioritize security",
			"flag SQL injection risks",
			"detect authorization issues",
			"detect secret leakage",
			"identify insecure file uploads",
		},
		Reviewers: []string{
			"security-engineer",
			"staff-backend-engineer",
		},
		Prompts: nil,
	}
}
