package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

	"wood-hot-monitor/ent"
	"wood-hot-monitor/ent/keywordexpansion"
	"wood-hot-monitor/internal/core/config"
	domain "wood-hot-monitor/internal/domain/hotspot"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

const defaultModel = "Pro/deepseek-ai/DeepSeek-V3.2"

type PreMatchResult struct {
	Matched      bool
	MatchedTerms []string
}

var (
	reArray  = regexp.MustCompile(`(?s)\[.*\]`)
	reObject = regexp.MustCompile(`(?s)\{.*\}`)

	validImportances = map[domain.Importance]bool{
		domain.ImportanceLow:    true,
		domain.ImportanceMedium: true,
		domain.ImportanceHigh:   true,
		domain.ImportanceUrgent: true,
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
	case s.semaphore <- struct{}{}:
		defer func() { <-s.semaphore }()
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

func PreMatchKeyword(text string, expandedKeywords []string) PreMatchResult {
	lower := strings.ToLower(text)
	var matched []string
	for _, kw := range expandedKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			matched = append(matched, kw)
		}
	}
	return PreMatchResult{Matched: len(matched) > 0, MatchedTerms: matched}
}

func (s *Service) AnalyzeContent(ctx context.Context, content, keyword string, expandedKeywords []string) (*domain.AnalysisResult, error) {
	var pm *PreMatchResult
	if len(expandedKeywords) > 0 {
		m := PreMatchKeyword(content, expandedKeywords)
		pm = &m
	}

	match := PreMatchResult{}
	if pm != nil {
		match = *pm
	}

	cfg, err := s.cfg.Get()
	if err != nil {
		return fallbackAnalysis(match, content), nil
	}
	if cfg.LLMAPIKey == "" {
		return fallbackAnalysis(match, content), nil
	}

	system, user := buildAnalysisPrompt(keyword, match, content)
	respText, err := s.callLLM(ctx, system, user)
	if err != nil {
		log.Printf("AI analysis failed: %v", err)
		return fallbackAnalysisError(match, content), nil
	}

	jsonStr := reObject.FindString(respText)
	if jsonStr == "" {
		return fallbackAnalysisError(match, content), nil
	}

	var raw struct {
		IsReal           bool             `json:"isReal"`
		Relevance        int              `json:"relevance"`
		RelevanceReason  string           `json:"relevanceReason"`
		KeywordMentioned bool             `json:"keywordMentioned"`
		Importance       domain.Importance `json:"importance"`
		Summary          string           `json:"summary"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return fallbackAnalysisError(match, content), nil
	}

	result := &domain.AnalysisResult{
		IsReal:           raw.IsReal,
		Relevance:        max(0, min(raw.Relevance, 100)),
		RelevanceReason:  truncateRunes(raw.RelevanceReason, 200),
		KeywordMentioned: raw.KeywordMentioned,
		Importance:       raw.Importance,
		Summary:          truncateRunes(raw.Summary, 150),
	}

	if !validImportances[result.Importance] {
		result.Importance = domain.ImportanceLow
	}

	return result, nil
}

func (s *Service) BatchAnalyze(ctx context.Context, contents []string, keyword string, expandedKeywords []string) ([]*domain.AnalysisResult, error) {
	results := make([]*domain.AnalysisResult, len(contents))
	var wg sync.WaitGroup

	for i, content := range contents {
		wg.Add(1)
		go func(idx int, c string) {
			defer wg.Done()
			r, _ := s.AnalyzeContent(ctx, c, keyword, expandedKeywords)
			results[idx] = r
		}(i, content)
	}

	wg.Wait()
	return results, nil
}

func (s *Service) ExpandKeyword(ctx context.Context, keyword string) ([]string, error) {
	cached, err := s.client.KeywordExpansion.Query().
		Where(keywordexpansion.KeywordEQ(keyword)).
		Select(keywordexpansion.FieldExpansion).
		Strings(ctx)
	if err != nil {
		return nil, fmt.Errorf("get cached expansions: %w", err)
	}
	if len(cached) > 0 {
		return dedupWithLeader(keyword, cached), nil
	}

	coreTerms := extractCoreTerms(keyword)

	system, user := buildExpandKeywordPrompt(keyword)
	respText, err := s.callLLM(ctx, system, user)
	if err != nil {
		log.Printf("keyword expansion LLM call failed: %v", err)
	}

	result := dedupWithLeader(keyword, coreTerms)

	if respText != "" {
		if jsonStr := reArray.FindString(respText); jsonStr != "" {
			var parsed []string
			if err := json.Unmarshal([]byte(jsonStr), &parsed); err == nil {
				seen := make(map[string]bool, len(result))
				for _, r := range result {
					seen[r] = true
				}
				for _, item := range parsed {
					item = strings.TrimSpace(item)
					if item != "" && !seen[item] {
						result = append(result, item)
						seen[item] = true
					}
				}
			}
		}
	}

	coreSet := make(map[string]bool, len(coreTerms))
	for _, t := range coreTerms {
		coreSet[t] = true
	}
	for _, exp := range result[1:] {
		if coreSet[exp] {
			continue
		}
		_, insertErr := s.client.KeywordExpansion.Create().
			SetID(uuid.NewString()).
			SetKeyword(keyword).
			SetExpansion(exp).
			Save(ctx)
		if insertErr != nil {
			log.Printf("cache expansion %q: %v", exp, insertErr)
		}
	}

	log.Printf("query expansion for %q: %d variants: %v", keyword, len(result), result)
	return result, nil
}

func extractCoreTerms(keyword string) []string {
	parts := splitTerms(keyword)
	if len(parts) <= 1 {
		return nil
	}

	seen := make(map[string]bool)
	lowerKw := strings.ToLower(keyword)
	var terms []string

	for _, p := range parts {
		if !seen[p] && strings.ToLower(p) != lowerKw {
			terms = append(terms, p)
			seen[p] = true
		}
	}

	for i := 0; i < len(parts)-1; i++ {
		bigram := parts[i] + " " + parts[i+1]
		if !seen[bigram] && strings.ToLower(bigram) != lowerKw {
			terms = append(terms, bigram)
			seen[bigram] = true
		}
	}
	return terms
}

func splitTerms(s string) []string {
	var parts []string
	for _, p := range strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == '-' || r == '_' || r == '/' || r == '\\' || r == '·'
	}) {
		if len([]rune(p)) >= 2 {
			parts = append(parts, p)
		}
	}
	return parts
}

func dedupWithLeader(leader string, items []string) []string {
	result := []string{leader}
	seen := map[string]bool{leader: true}
	for _, item := range items {
		if !seen[item] {
			result = append(result, item)
			seen[item] = true
		}
	}
	return result
}

func fallbackAnalysis(match PreMatchResult, content string) *domain.AnalysisResult {
	relevance := 20
	if match.Matched {
		relevance = 50
	}
	return &domain.AnalysisResult{
		IsReal:           true,
		Relevance:        relevance,
		RelevanceReason:  "未配置 AI 服务，使用默认分数",
		KeywordMentioned: match.Matched,
		Importance:       domain.ImportanceLow,
		Summary:          truncateRunes(content, 50) + "...",
	}
}

func fallbackAnalysisError(match PreMatchResult, content string) *domain.AnalysisResult {
	relevance := 10
	if match.Matched {
		relevance = 30
	}
	return &domain.AnalysisResult{
		IsReal:           true,
		Relevance:        relevance,
		RelevanceReason:  "AI 分析失败，使用默认分数",
		KeywordMentioned: match.Matched,
		Importance:       domain.ImportanceLow,
		Summary:          truncateRunes(content, 50) + "...",
	}
}

func truncateRunes(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen])
	}
	return s
}
