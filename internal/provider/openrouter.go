package provider

const OpenRouterBaseURL = "https://openrouter.ai/api/v1"

func NewOpenRouterProvider(apiKey string, model string, appURL string, appTitle string) OpenAICompatibleProvider {
	headers := map[string]string{
		"HTTP-Referer": appURL,
		"X-Title":      appTitle,
	}

	return NewOpenAICompatibleProvider(OpenAICompatibleConfig{
		BaseURL: OpenRouterBaseURL,
		APIKey:  apiKey,
		Model:   model,
		Headers: headers,
	})
}
