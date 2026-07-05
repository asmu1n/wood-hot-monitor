package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"wood-hot-monitor/ent/keywordexpansion"

	"github.com/google/uuid"
)

// ExpandKeyword 将关键词扩展为多个检索变体（含原词）。
// 优先读取 DB 缓存；未命中时先提取核心词，再调用 AI 扩展，最后将 AI 结果写入缓存。
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
