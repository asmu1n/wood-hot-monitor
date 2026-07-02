package quality

import (
	"math"
	"sort"
	"time"

	"wood-hot-monitor/internal/infra/scraper"
)

// 质量评分三维度权重配置（按平台区分）
type PlatformWeights struct {
	Engagement float64 // 互动指标权重
	Authority  float64 // 作者权威度权重
	Recency    float64 // 时效性权重
}

// 平台最低质量阈值
type PlatformThreshold struct {
	MinScore float64
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

var platformThresholds = map[string]PlatformThreshold{
	"twitter":    {MinScore: 15},
	"weibo":      {MinScore: 15},
	"bilibili":   {MinScore: 10},
	"hackernews": {MinScore: 10},
	"bing":       {MinScore: 5},
	"sogou":      {MinScore: 5},
	"google":     {MinScore: 5},
	"duckduckgo": {MinScore: 5},
}

var defaultWeights = PlatformWeights{Engagement: 0.3, Authority: 0.3, Recency: 0.4}
var defaultThreshold = PlatformThreshold{MinScore: 5}

// ScoredResult 附带质量分数的搜索结果
type ScoredResult struct {
	scraper.SearchResult
	QualityScore float64
}

// ScoreResult 对单条结果计算质量分数（0~100）
func ScoreResult(r scraper.SearchResult) float64 {
	w := defaultWeights
	if pw, ok := platformWeights[r.Source]; ok {
		w = pw
	}

	eng := engagementScore(r)
	auth := authorityScore(r)
	rec := recencyScore(r)

	return clamp(eng*w.Engagement+auth*w.Authority+rec*w.Recency, 0, 100)
}

// FilterAndSort 过滤低于平台阈值的结果，按质量分降序排列
func FilterAndSort(results []scraper.SearchResult) []ScoredResult {
	var scored []ScoredResult
	for _, r := range results {
		score := ScoreResult(r)
		threshold := defaultThreshold
		if t, ok := platformThresholds[r.Source]; ok {
			threshold = t
		}
		if score >= threshold.MinScore {
			scored = append(scored, ScoredResult{SearchResult: r, QualityScore: score})
		}
	}
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].QualityScore > scored[j].QualityScore
	})
	return scored
}

// engagementScore 计算互动指标得分（0~100）
// 综合点赞、评论、转发、播放等指标，使用对数缩放
func engagementScore(r scraper.SearchResult) float64 {
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
	// 对数缩放：log10(total+1) * 20，上限100
	return clamp(math.Log10(total+1)*20, 0, 100)
}

// authorityScore 计算作者权威度得分（0~100）
func authorityScore(r scraper.SearchResult) float64 {
	if r.Author == nil {
		return 20
	}
	score := 20.0
	if r.Author.Verified {
		score += 30
	}
	if r.Author.Followers > 0 {
		// 粉丝数对数缩放
		score += clamp(math.Log10(float64(r.Author.Followers))*10, 0, 50)
	}
	return clamp(score, 0, 100)
}

// recencyScore 计算时效性得分（0~100），基于时间衰减
func recencyScore(r scraper.SearchResult) float64 {
	if r.PublishedAt == nil {
		return 30
	}
	hours := time.Since(*r.PublishedAt).Hours()
	if hours < 0 {
		hours = 0
	}
	// 指数衰减：1h=100, 24h≈60, 72h≈30, 168h(7d)≈10
	return clamp(100*math.Exp(-hours/72), 0, 100)
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
