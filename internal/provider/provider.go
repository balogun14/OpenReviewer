package provider

type Request struct {
	Owner               string
	Repository          string
	PullRequestNumber   int
	PullRequestTitle    string
	PullRequestBody     string
	Diff                string
	ProfileRules        []string
	PersonaID           string
	PersonaDisplayName  string
	PersonaInstructions []string
	PromptID            string
	PromptName          string
	Prompt              string
}
