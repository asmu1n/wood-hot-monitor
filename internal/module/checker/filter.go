package checker

import (
	"math"
	"sort"
	"time"
	"wood-hot-monitor/internal/module/hotspot"
)

var sourcePriority = map[string]int{
	"twitter":    0,
	"weibo":      1,
	"bilibili":   2,
	"hackernews": 3,
	"sogou":      4,
	"bing":       5,
	"google":     6,
	"duckduckgo": 7,
}

// 根据URL去重
func deduplicateByURL(results []hotspot.SearchResult) []hotspot.SearchResult {
	seen := make(map[string]bool, len(results))
	out := make([]hotspot.SearchResult, 0, len(results))
	for _, r := range results {
		if r.URL == "" || seen[r.URL] {
			continue
		}
		seen[r.URL] = true
		out = append(out, r)
	}
	return out
}

// 根据时间过滤
func filterByFreshness(results []hotspot.SearchResult, maxAge time.Duration) []hotspot.SearchResult {
	cutoff := time.Now().Add(-maxAge)
	out := make([]hotspot.SearchResult, 0, len(results))
	for _, r := range results {
		if r.PublishedAt == nil || r.PublishedAt.After(cutoff) {
			out = append(out, r)
		}
	}
	return out
}

// 根据信息来源权重排序
func sortByPriority(results []hotspot.SearchResult) {
	sort.SliceStable(results, func(i, j int) bool {
		pi := priorityOf(results[i].Source)
		pj := priorityOf(results[j].Source)
		return pi < pj
	})
}

func priorityOf(source string) int {
	if p, ok := sourcePriority[source]; ok {
		return p
	}
	return 99
}

// PlatformWeights 平台权重配置
type PlatformWeights struct {
	Engagement float64
	Authority  float64
	Recency    float64
}

type ScoredResult struct {
	hotspot.SearchResult
	QualityScore float64
}

var platformWeights = map[string]PlatformWeights{
	"twitter":    {Engagement: 0.4, Authority: 0.35, Recency: 0.25},
	"weibo":      {Engagement: 0.4, Authority: 0.3, Recency: 0.3},
	"bilibili":   {Engagement: 0.45, Authority: 0.25, Recency: 0.3},
	"hackernews": {Engagement: 0.5, Authority: 0.2, Recency: 0.3},
	"bing":       {Engagement: 0.2, Authority: 0.3, Recency: 0.5},
	"sogou":      {Engagement: 0.2, Authority: 0.3, Recency: 0.5},
	"google":     {Engagement: 0.2, Authority: 0.3, Recency: 0.5},
	"duckduckgo": {Engagement: 0.2, Authority: 0.3, Recency: 0.5},
}

var platformThresholds = map[string]float64{
	"twitter":    15,
	"weibo":      15,
	"bilibili":   10,
	"hackernews": 10,
	"bing":       5,
	"sogou":      5,
	"google":     5,
	"duckduckgo": 5,
}

var defaultWeights = PlatformWeights{Engagement: 0.3, Authority: 0.3, Recency: 0.4}

func scoreResult(r hotspot.SearchResult) float64 {
	w := defaultWeights
	if pw, ok := platformWeights[r.Source]; ok {
		w = pw
	}
	eng := engagementScore(r)
	auth := authorityScore(r)
	rec := recencyScore(r)
	return max(0, min(eng*w.Engagement+auth*w.Authority+rec*w.Recency, 100))
}

func filterAndSort(results []hotspot.SearchResult) []ScoredResult {
	var scored []ScoredResult
	for _, r := range results {
		score := scoreResult(r)
		threshold := 5.0
		if t, ok := platformThresholds[r.Source]; ok {
			threshold = t
		}
		if score >= threshold {
			scored = append(scored, ScoredResult{SearchResult: r, QualityScore: score})
		}
	}
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].QualityScore > scored[j].QualityScore
	})
	return scored
}

func engagementScore(r hotspot.SearchResult) float64 {
	total := 0.0
	if r.LikeCount != nil {
		total += float64(*r.LikeCount) * 1.0
	}
	if r.CommentCount != nil {
		total += float64(*r.CommentCount) * 2.0
	}
	if r.RetweetCount != nil {
		total += float64(*r.RetweetCount) * 1.5
	}
	if r.ReplyCount != nil {
		total += float64(*r.ReplyCount) * 2.0
	}
	if r.QuoteCount != nil {
		total += float64(*r.QuoteCount) * 1.5
	}
	if r.ViewCount != nil {
		total += float64(*r.ViewCount) * 0.01
	}
	if r.DanmakuCount != nil {
		total += float64(*r.DanmakuCount) * 1.0
	}
	if total <= 0 {
		return 10
	}
	return max(0.0, min(math.Log10(total+1)*20, 100))
}

func authorityScore(r hotspot.SearchResult) float64 {
	if r.Author == nil {
		return 20
	}
	score := 20.0
	if r.Author.Verified {
		score += 30
	}
	if r.Author.Followers > 0 {
		score += max(0.0, min(math.Log10(float64(r.Author.Followers))*10, 50))
	}
	return max(0.0, min(score, 100))
}

func recencyScore(r hotspot.SearchResult) float64 {
	if r.PublishedAt == nil {
		return 30
	}
	hours := time.Since(*r.PublishedAt).Hours()
	if hours < 0 {
		hours = 0
	}
	return max(0.0, min(100*math.Exp(-hours/72), 100))
}
