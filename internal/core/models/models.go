package models

import (
	"time"
)

type Keyword struct {
	ID           string  `json:"id"`
	Text         string  `json:"text"`
	Category     *string `json:"category"`
	IsActive     bool    `json:"isActive"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
	HotspotCount *int    `json:"hotspotCount,omitempty"`
}

type Hotspot struct {
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	Content          string          `json:"content"`
	URL              string          `json:"url"`
	Source           string          `json:"source"`
	SourceID         *string         `json:"sourceId"`
	IsReal           bool            `json:"isReal"`
	Relevance        int             `json:"relevance"`
	RelevanceReason  *string         `json:"relevanceReason"`
	KeywordMentioned *bool           `json:"keywordMentioned"`
	Importance       Importance      `json:"importance"`
	Summary          *string         `json:"summary"`
	ViewCount        *int            `json:"viewCount"`
	LikeCount        *int            `json:"likeCount"`
	RetweetCount     *int            `json:"retweetCount"`
	ReplyCount       *int            `json:"replyCount"`
	CommentCount     *int            `json:"commentCount"`
	QuoteCount       *int            `json:"quoteCount"`
	DanmakuCount     *int            `json:"danmakuCount"`
	AuthorName       *string         `json:"authorName"`
	AuthorUsername   *string         `json:"authorUsername"`
	AuthorAvatar     *string         `json:"authorAvatar"`
	AuthorFollowers  *int            `json:"authorFollowers"`
	AuthorVerified   *bool           `json:"authorVerified"`
	PublishedAt      *string         `json:"publishedAt"`
	CreatedAt        string          `json:"createdAt"`
	KeywordID        *string         `json:"-"`
	Keyword          *HotspotKeyword `json:"keyword"`
	IsNotified       bool            `json:"isNotified"`
	NotifiedAt       *string         `json:"notifiedAt"`
	IsRead           bool            `json:"isRead"`
}

type HotspotKeyword struct {
	ID       string  `json:"id"`
	Text     string  `json:"text"`
	Category *string `json:"category"`
}

type Status struct {
	Total    int            `json:"total"`
	Today    int            `json:"today"`
	Urgent   int            `json:"urgent"`
	BySource map[string]int `json:"bySource"`
}

type PaginatedResult[T any] struct {
	Data  []T `json:"data"`
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

type AppConfig struct {
	EmailAddress  string         `json:"emailAddress"`
	LLMModel      string         `json:"llmModel"`
	LLMAPIKey     string         `json:"llmApiKey"`
	LLMBaseURL    string         `json:"llmBaseUrl"`
	CheckInterval int            `json:"checkInterval"`
	Settings      map[string]any `json:"settings"`
}

type KeywordExpansion struct {
	ID        string `json:"id"`
	Keyword   string `json:"keyword"`
	Expansion string `json:"expansion"`
	CreatedAt string `json:"createdAt"`
}

// SearchResult 表示从任意来源抓取到的一条搜索结果
type SearchResult struct {
	Title    string  `json:"title"`
	Content  string  `json:"content"`
	URL      string  `json:"url"`
	Source   string  `json:"source"`
	SourceID string  `json:"sourceId"`
	Author   *Author `json:"author"`

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

// AnalysisResult 表示 LLM 内容分析的结构化结果
type AnalysisResult struct {
	IsReal           bool       `json:"isReal"`
	Relevance        int        `json:"relevance"`
	RelevanceReason  string     `json:"relevanceReason"`
	KeywordMentioned bool       `json:"keywordMentioned"`
	Importance       Importance `json:"importance"`
	Summary          string     `json:"summary"`
}

type Importance string

const (
	LowImportance    Importance = "low"
	MediumImportance Importance = "medium"
	HighImportance   Importance = "high"
	UrgentImportance Importance = "urgent"
)

func (i Importance) Values() []string {
	return []string{string(LowImportance), string(MediumImportance), string(HighImportance), string(UrgentImportance)}
}
