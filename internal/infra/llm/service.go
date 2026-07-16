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
	"wood-hot-monitor/internal/config"
	"wood-hot-monitor/internal/domain/hotspot"

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

	validImportances = map[hotspot.Importance]bool{
		hotspot.ImportanceLow:    true,
		hotspot.ImportanceMedium: true,
		hotspot.ImportanceHigh:   true,
		hotspot.ImportanceUrgent: true,
	}
)

type Service struct {
	cfg       *config.Service
	client    *ent.Client
	semaphore chan struct{}
}

// 通过 `semaphore` 信道 设置 `30` 的并发限制
func NewService(cfg *config.Service, client *ent.Client) *Service {
	return &Service{
		cfg:       cfg,
		client:    client,
		semaphore: make(chan struct{}, 30),
	}
}

// 调用 LLM 服务
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

	// select 多路判断实现并发限制（如果任务上下文还未结束时，会持续阻塞等待）
	select {
	case s.semaphore <- struct{}{}:
		// 再判断一次 ctx 是否正常，避免两种 case 同时满足时随机进入正常分支
		if err := ctx.Err(); err != nil {
			<-s.semaphore // 释放刚才抢到的名额
			return "", err
		}
		defer func() { <-s.semaphore }()

	case <-ctx.Done():
		return "", ctx.Err()
	}

	// 调用 LLM 服务，并设定参数
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

func (s *Service) AnalyzeContent(ctx context.Context, content, keyword string, expandedKeywords []string) (*hotspot.AnalysisResult, error) {
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
		IsReal           bool               `json:"isReal"`
		Relevance        int                `json:"relevance"`
		RelevanceReason  string             `json:"relevanceReason"`
		KeywordMentioned bool               `json:"keywordMentioned"`
		Importance       hotspot.Importance `json:"importance"`
		Summary          string             `json:"summary"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return fallbackAnalysisError(match, content), nil
	}

	result := &hotspot.AnalysisResult{
		IsReal:           raw.IsReal,
		Relevance:        max(0, min(raw.Relevance, 100)),
		RelevanceReason:  truncateRunes(raw.RelevanceReason, 200),
		KeywordMentioned: raw.KeywordMentioned,
		Importance:       raw.Importance,
		Summary:          truncateRunes(raw.Summary, 150),
	}

	if !validImportances[result.Importance] {
		result.Importance = hotspot.ImportanceLow
	}

	return result, nil
}

func (s *Service) BatchAnalyze(ctx context.Context, contents []string, keyword string, expandedKeywords []string) ([]*hotspot.AnalysisResult, error) {
	results := make([]*hotspot.AnalysisResult, len(contents))
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
			parsed := make([]string, 0, 5)
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

func fallbackAnalysis(match PreMatchResult, content string) *hotspot.AnalysisResult {
	relevance := 20
	if match.Matched {
		relevance = 50
	}
	return &hotspot.AnalysisResult{
		IsReal:           true,
		Relevance:        relevance,
		RelevanceReason:  "未配置 AI 服务，使用默认分数",
		KeywordMentioned: match.Matched,
		Importance:       hotspot.ImportanceLow,
		Summary:          truncateRunes(content, 50) + "...",
	}
}

func fallbackAnalysisError(match PreMatchResult, content string) *hotspot.AnalysisResult {
	relevance := 10
	if match.Matched {
		relevance = 30
	}
	return &hotspot.AnalysisResult{
		IsReal:           true,
		Relevance:        relevance,
		RelevanceReason:  "AI 分析失败，使用默认分数",
		KeywordMentioned: match.Matched,
		Importance:       hotspot.ImportanceLow,
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
