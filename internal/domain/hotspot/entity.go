package hotspot

import "time"

// Hotspot 热点聚合根
type Hotspot struct {
	ID               string
	Title            string
	Content          string
	URL              string
	Source           string
	SourceID         *string
	IsReal           bool
	Relevance        int
	RelevanceReason  *string
	KeywordMentioned *bool
	Importance       Importance
	Summary          *string
	ViewCount    *int
	LikeCount    *int
	RetweetCount *int
	ReplyCount   *int
	CommentCount *int
	QuoteCount   *int
	DanmakuCount *int
	Author       *Author
	PublishedAt  *time.Time
	CreatedAt        time.Time
	KeywordID        *string
	Keyword          *HotspotKeyword
	IsNotified       bool
	NotifiedAt       *time.Time
	IsRead           bool
}

// HotspotKeyword 热点关联的关键词（简化视图）
type HotspotKeyword struct {
	ID       string
	Text     string
	Category *string
}

// Author 内容发布者信息
type Author struct {
	Name      string
	Username  string
	Avatar    string
	Followers int
	Verified  bool
}

// SearchResult 从任意来源抓取到的一条搜索结果
type SearchResult struct {
	Title    string
	Content  string
	URL      string
	Source   string
	SourceID string
	Author   *Author

	ViewCount    *int
	LikeCount    *int
	RetweetCount *int
	ReplyCount   *int
	CommentCount *int
	QuoteCount   *int
	DanmakuCount *int

	PublishedAt *time.Time
}

// AnalysisResult LLM 内容分析的结构化结果
type AnalysisResult struct {
	IsReal           bool       `json:"isReal"`
	Relevance        int        `json:"relevance"`
	RelevanceReason  string     `json:"relevanceReason"`
	KeywordMentioned bool       `json:"keywordMentioned"`
	Importance       Importance `json:"importance"`
	Summary          string     `json:"summary"`
}

// Status 热点统计概览
type Status struct {
	Total    int
	Today    int
	Urgent   int
	BySource map[string]int
}
