package config

import "wood-hot-monitor/pkg/types"

type LLMConfig struct {
	LLMModel   string `json:"llmModel"`
	LLMAPIKey  string `json:"llmApiKey"`
	LLMBaseURL string `json:"llmBaseUrl"`
}

type NotifyConfig struct {
	OSNotifyEnabled       bool             `json:"osNotifyEnabled"`
	OSNotifyMinImportance types.Importance `json:"osNotifyMinImportance"`
	EmailAddress          string           `json:"emailAddress"`
	ResendApiKey          string           `json:"resendApiKey"`
}

type CheckConfig struct {
	CheckInterval int    `json:"checkInterval"`
	TwitterApiKey string `json:"twitterApiKey"`
}

type AppConfig struct {
	LLMConfig
	NotifyConfig
	CheckConfig
}
