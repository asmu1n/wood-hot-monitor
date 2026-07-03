package llm

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"
)

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

	result.Relevance = max(0, min(result.Relevance, 100))
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

// 截断过长的文本
func truncateRunes(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen])
	}
	return s
}
