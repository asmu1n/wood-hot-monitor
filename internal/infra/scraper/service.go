package scraper

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"sync"
	"time"

	"wood-hot-monitor/internal/domain/hotspot"
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

var userAgents = [4]string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:133.0) Gecko/20100101 Firefox/133.0",
}

func RandomUA() string {
	return userAgents[rand.Intn(len(userAgents))]
}

func NewHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}

// 模拟浏览器请求头
func SetRequestHeaders(req *http.Request) {
	req.Header.Set("User-Agent", RandomUA())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) SearchAll(ctx context.Context, query string, config hotspot.ScraperConfig) []hotspot.SearchResult {
	type sourceResult struct {
		results []hotspot.SearchResult
		source  string
		err     error
	}

	type searchTask struct {
		name string
		fn   func() ([]hotspot.SearchResult, error)
	}

	// 定义搜索任务
	tasks := []searchTask{
		{"hackernews", func() ([]hotspot.SearchResult, error) { return SearchHackerNews(ctx, query) }},
		{"bing", func() ([]hotspot.SearchResult, error) { return SearchBing(ctx, query) }},
		{"bilibili", func() ([]hotspot.SearchResult, error) { return SearchBilibili(ctx, query) }},
		{"twitter", func() ([]hotspot.SearchResult, error) { return SearchTwitter(ctx, query, config.TwitterAPIKey) }},
	}

	// 创建通道和等待组
	ch := make(chan sourceResult, len(tasks))
	var wg sync.WaitGroup

	// 启动所有搜索任务
	for _, t := range tasks {
		wg.Add(1)

		go func(name string, fn func() ([]hotspot.SearchResult, error)) {
			defer wg.Done()

			// 错误捕获兜底
			defer func() {
				if r := recover(); r != nil {
					ch <- sourceResult{
						source: name,
						err:    fmt.Errorf("panic recovered: %v", r),
					}
				}
			}()
			results, err := fn()
			ch <- sourceResult{
				results: results,
				source:  name,
				err:     err,
			}
		}(t.name, t.fn)
	}

	// 等待任务完成然后关闭信道，让接收端处理全部结果
	go func() {
		wg.Wait()
		close(ch)
	}()

	all := make([]hotspot.SearchResult, 20)

	// for select 持续尝试接收任务结果，并且在ctx取消时返回已收集的结果
	for {
		select {
		case sr, ok := <-ch:
			if !ok {
				return all
			}
			if sr.err != nil {
				log.Printf("scraper: %s search failed: %v", sr.source, sr.err)
				continue
			}
			all = append(all, sr.results...)

		case <-ctx.Done():
			return all
		}
	}

}

// 根据URL去重
func DeduplicateByURL(results []hotspot.SearchResult) []hotspot.SearchResult {
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
func FilterByFreshness(results []hotspot.SearchResult, maxAge time.Duration) []hotspot.SearchResult {
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
func SortByPriority(results []hotspot.SearchResult) {
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

func ScoreResult(r hotspot.SearchResult) float64 {
	w := defaultWeights
	if pw, ok := platformWeights[r.Source]; ok {
		w = pw
	}
	eng := engagementScore(r)
	auth := authorityScore(r)
	rec := recencyScore(r)
	return max(0, min(eng*w.Engagement+auth*w.Authority+rec*w.Recency, 100))
}

func FilterAndSort(results []hotspot.SearchResult) []ScoredResult {
	var scored []ScoredResult
	for _, r := range results {
		score := ScoreResult(r)
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
