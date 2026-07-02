package scraper

import (
	"math/rand"
	"net/http"
	"sort"
	"time"
)

// SearchResult 表示从任意来源抓取到的一条搜索结果
type SearchResult struct {
	Title     string  `json:"title"`
	Content   string  `json:"content"`
	URL       string  `json:"url"`
	Source    string  `json:"source"`
	SourceID  string  `json:"sourceId"`
	Author    *Author `json:"author"`

	// 互动指标（各平台按需填充）
	ViewCount    *int `json:"viewCount"`
	LikeCount    *int `json:"likeCount"`
	RetweetCount *int `json:"retweetCount"`
	ReplyCount   *int `json:"replyCount"`
	CommentCount *int `json:"commentCount"`
	QuoteCount   *int `json:"quoteCount"`
	DanmakuCount *int `json:"danmakuCount"`

	PublishedAt *time.Time `json:"publishedAt"`
}

// Author 表示内容发布者信息
type Author struct {
	Name      string `json:"name"`
	Username  string `json:"username"`
	Avatar    string `json:"avatar"`
	Followers int    `json:"followers"`
	Verified  bool   `json:"verified"`
}

// sourcePriority 平台优先级映射，数值越小越优先
// 参考 hotspotCheck.ts: twitter > weibo > bilibili > hackernews > sogou > bing > google > duckduckgo
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

// userAgents 随机 User-Agent 池，模拟真实浏览器请求
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:133.0) Gecko/20100101 Firefox/133.0",
}

// RandomUA 返回随机 User-Agent 字符串
func RandomUA() string {
	return userAgents[rand.Intn(len(userAgents))]
}

// NewHTTPClient 创建带超时和随机 UA 的 HTTP 客户端
func NewHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}

// SetRequestHeaders 为请求设置通用浏览器头
func SetRequestHeaders(req *http.Request) {
	req.Header.Set("User-Agent", RandomUA())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
}

// DeduplicateByURL 按 URL 去重，保留首次出现的结果
func DeduplicateByURL(results []SearchResult) []SearchResult {
	seen := make(map[string]bool, len(results))
	out := make([]SearchResult, 0, len(results))
	for _, r := range results {
		if r.URL == "" || seen[r.URL] {
			continue
		}
		seen[r.URL] = true
		out = append(out, r)
	}
	return out
}

// FilterByFreshness 过滤掉超过 maxAge 的结果（基于 PublishedAt）
// 没有 PublishedAt 的结果会被保留（无法判断时效性）
func FilterByFreshness(results []SearchResult, maxAge time.Duration) []SearchResult {
	cutoff := time.Now().Add(-maxAge)
	out := make([]SearchResult, 0, len(results))
	for _, r := range results {
		if r.PublishedAt == nil || r.PublishedAt.After(cutoff) {
			out = append(out, r)
		}
	}
	return out
}

// SortByPriority 按平台优先级排序（优先级相同时保留原始顺序）
func SortByPriority(results []SearchResult) {
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

// IntPtr 辅助函数，将 int 转为指针
func IntPtr(v int) *int { return &v }

// TimePtr 辅助函数，将 time.Time 转为指针
func TimePtr(t time.Time) *time.Time { return &t }
