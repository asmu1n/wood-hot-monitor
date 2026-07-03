package llm

import (
	"context"
	"fmt"
	"regexp"
	"wood-hot-monitor/ent"
	"wood-hot-monitor/internal/core/config"
	"wood-hot-monitor/internal/core/models"

	"github.com/sashabaranov/go-openai"
)

const defaultModel = "Pro/deepseek-ai/DeepSeek-V3.2"

type AnalysisResult = models.AnalysisResult

type PreMatchResult struct {
	Matched      bool
	MatchedTerms []string
}

var (
	reArray  = regexp.MustCompile(`(?s)\[.*\]`)
	reObject = regexp.MustCompile(`(?s)\{.*\}`)

	validImportances = map[string]bool{
		"low": true, "medium": true, "high": true, "urgent": true,
	}
)

type Service struct {
	cfg       *config.Service
	client    *ent.Client
	semaphore chan struct{}
}

func NewService(cfg *config.Service, client *ent.Client) *Service {
	return &Service{
		cfg:       cfg,
		client:    client,
		semaphore: make(chan struct{}, 30),
	}
}

// callLLM 向配置的 LLM 提供商发送 chat completion 请求，返回模型回复文本。
// 未配置 API Key 时返回空字符串（非错误）。通过 semaphore 控制并发上限。
func (s *Service) callLLM(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	cfg, err := s.cfg.Get()
	if err != nil {
		return "", fmt.Errorf("get config: %w", err)
	}
	if cfg.LLMAPIKey == "" {
		return "", nil
	}

	clientCfg := openai.DefaultConfig(cfg.LLMAPIKey)
	if cfg.LLMBaseURL != "" {
		clientCfg.BaseURL = cfg.LLMBaseURL
	}

	model := cfg.LLMModel
	if model == "" {
		model = defaultModel
	}

	client := openai.NewClientWithConfig(clientCfg)

	select {
	// 没触发并发限制的话，推入缓冲栈，在调用完成后弹出。若触发并发限制，则阻塞在这里
	case s.semaphore <- struct{}{}:
		defer func() { <-s.semaphore }()
	// 如果上下文超时，直接结束任务
	case <-ctx.Done():
		return "", ctx.Err()
	}

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userPrompt},
		},
		Temperature: 0.2,
		TopP:        0.95,
		MaxTokens:   1024,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	})
	if err != nil {
		return "", fmt.Errorf("chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", nil
	}
	return resp.Choices[0].Message.Content, nil
}
