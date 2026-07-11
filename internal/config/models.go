package config

type AppConfig struct {
	EmailAddress  string         `json:"emailAddress"`
	LLMModel      string         `json:"llmModel"`
	LLMAPIKey     string         `json:"llmApiKey"`
	LLMBaseURL    string         `json:"llmBaseUrl"`
	CheckInterval int            `json:"checkInterval"`
	Settings      map[string]any `json:"settings"`
}
