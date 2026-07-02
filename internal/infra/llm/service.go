package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"wood-hot-monitor/ent/keywordexpansion"
	"wood-hot-monitor/internal/core/config"
	"wood-hot-monitor/internal/infra/database"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

const defaultModel = "Pro/deepseek-ai/DeepSeek-V3.2"

type AnalysisResult struct {
	IsReal           bool   `json:"isReal"`
	Relevance        int    `json:"relevance"`
	RelevanceReason  string `json:"relevanceReason"`
	KeywordMentioned bool   `json:"keywordMentioned"`
	Importance       string `json:"importance"`
	Summary          string `json:"summary"`
}

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
	db        *database.DB
	semaphore chan struct{}
}

// NewService 创建 LLM 服务实例，semaphore 限制最大并发调用数为 10
func NewService(cfg *config.Service, db *database.DB) *Service {
	return &Service{
		cfg:       cfg,
		db:        db,
		semaphore: make(chan struct{}, 10),
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

// ExpandKeyword 将关键词扩展为多个检索变体（含原词）。
// 优先读取 DB 缓存；未命中时先提取核心词，再调用 AI 扩展，最后将 AI 结果写入缓存。
func (s *Service) ExpandKeyword(ctx context.Context, keyword string) ([]string, error) {
	cached, err := s.db.Client.KeywordExpansion.Query().
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
				for _, s := range parsed {
					s = strings.TrimSpace(s)
					if s != "" && !seen[s] {
						result = append(result, s)
						seen[s] = true
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
		_, insertErr := s.db.Client.KeywordExpansion.Create().
			SetID(uuid.NewString()).
			SetKeyword(keyword).
			SetExpansion(exp).
			Save(ctx)
		if insertErr != nil {
			log.Printf("cache expansion %q: %v", exp, insertErr)
		}
	}

	log.Printf("query expansion for %q: %d variants", keyword, len(result))
	return result, nil
}

// PreMatchKeyword 在文本中检查是否包含任一扩展关键词（不区分大小写），返回匹配结果。
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

// AnalyzeContent 使用 AI 分析内容与关键词的相关性，返回结构化评分结果。
func (s *Service) AnalyzeContent(ctx context.Context, content, keyword string, preMatch *PreMatchResult) (*AnalysisResult, error) {
	match := PreMatchResult{}
	if preMatch != nil {
		match = *preMatch
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

	var result AnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return fallbackAnalysisError(match, content), nil
	}

	result.Relevance = clamp(result.Relevance, 0, 100)
	if !validImportances[result.Importance] {
		result.Importance = "low"
	}
	result.RelevanceReason = truncateRunes(result.RelevanceReason, 200)
	result.Summary = truncateRunes(result.Summary, 150)

	return &result, nil
}

// BatchAnalyze 并发分析多条内容，每条自动执行预匹配后调用 AnalyzeContent。
func (s *Service) BatchAnalyze(ctx context.Context, contents []string, keyword string, expandedKeywords []string) ([]*AnalysisResult, error) {
	results := make([]*AnalysisResult, len(contents))
	var wg sync.WaitGroup

	for i, content := range contents {
		wg.Add(1)
		go func(idx int, c string) {
			defer wg.Done()
			var pm *PreMatchResult
			if len(expandedKeywords) > 0 {
				m := PreMatchKeyword(c, expandedKeywords)
				pm = &m
			}
			r, _ := s.AnalyzeContent(ctx, c, keyword, pm)
			results[idx] = r
		}(i, content)
	}

	wg.Wait()
	return results, nil
}

// extractCoreTerms 从组合关键词中提取核心子词及相邻二元组合
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

// splitTerms 按空格、连字符、下划线等分隔符拆分字符串，过滤长度 < 2 的片段
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

// dedupWithLeader 以 leader 为首元素，将 items 去重追加
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

// fallbackAnalysis 未配置 AI 时的降级评分
func fallbackAnalysis(match PreMatchResult, content string) *AnalysisResult {
	relevance := 20
	if match.Matched {
		relevance = 50
	}
	return &AnalysisResult{
		IsReal:           true,
		Relevance:        relevance,
		RelevanceReason:  "未配置 AI 服务，使用默认分数",
		KeywordMentioned: match.Matched,
		Importance:       "low",
		Summary:          truncateRunes(content, 50) + "...",
	}
}

// fallbackAnalysisError AI 调用失败时的降级评分
func fallbackAnalysisError(match PreMatchResult, content string) *AnalysisResult {
	relevance := 10
	if match.Matched {
		relevance = 30
	}
	return &AnalysisResult{
		IsReal:           true,
		Relevance:        relevance,
		RelevanceReason:  "AI 分析失败，使用默认分数",
		KeywordMentioned: match.Matched,
		Importance:       "low",
		Summary:          truncateRunes(content, 50) + "...",
	}
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func truncateRunes(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen])
	}
	return s
}
