package config

import "wood-hot-monitor/pkg/types"

type AppConfig struct {
	EmailAddress  string         `json:"emailAddress"`
	LLMModel      string         `json:"llmModel"`
	LLMAPIKey     string         `json:"llmApiKey"`
	LLMBaseURL    string         `json:"llmBaseUrl"`
	CheckInterval int            `json:"checkInterval"`
	Settings      map[string]any `json:"settings"`

	// OS 原生通知（系统通知中心 / Toast）。
	OSNotifyEnabled       bool             `json:"osNotifyEnabled"`
	OSNotifyMinImportance types.Importance `json:"osNotifyMinImportance"` // low | medium | high | urgent
}
