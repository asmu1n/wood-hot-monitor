package hotspot

import "time"

type Hotspot struct {
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	Content          string          `json:"content"`
	URL              string          `json:"url"`
	Source           string          `json:"source"`
	SourceID         *string         `json:"sourceId,omitempty"`
	IsReal           bool            `json:"isReal"`
	Relevance        int             `json:"relevance"`
	RelevanceReason  *string         `json:"relevanceReason,omitempty"`
	KeywordMentioned *bool           `json:"keywordMentioned,omitempty"`
	Importance       Importance      `json:"importance"`
	Summary          *string         `json:"summary,omitempty"`
	ViewCount        *int            `json:"viewCount,omitempty"`
	LikeCount        *int            `json:"likeCount,omitempty"`
	RetweetCount     *int            `json:"retweetCount,omitempty"`
	ReplyCount       *int            `json:"replyCount,omitempty"`
	CommentCount     *int            `json:"commentCount,omitempty"`
	QuoteCount       *int            `json:"quoteCount,omitempty"`
	DanmakuCount     *int            `json:"danmakuCount,omitempty"`
	Author           *Author         `json:"author,omitempty"`
	PublishedAt      *time.Time      `json:"publishedAt,omitempty"`
	CreatedAt        time.Time       `json:"createdAt"`
	KeywordID        *string         `json:"keywordId,omitempty"`
	Keyword          *HotspotKeyword `json:"keyword,omitempty"`
	IsNotified       bool            `json:"isNotified"`
	NotifiedAt       *time.Time      `json:"notifiedAt,omitempty"`
	IsRead           bool            `json:"isRead"`
}

type HotspotKeyword struct {
	ID       string  `json:"id"`
	Text     string  `json:"text"`
	Category *string `json:"category,omitempty"`
}

type Author struct {
	Name      string `json:"name"`
	Username  string `json:"username"`
	Avatar    string `json:"avatar"`
	Followers int    `json:"followers"`
	Verified  bool   `json:"verified"`
}

type SearchResult struct {
	Title    string  `json:"title"`
	Content  string  `json:"content"`
	URL      string  `json:"url"`
	Source   string  `json:"source"`
	SourceID string  `json:"sourceId"`
	Author   *Author `json:"author,omitempty"`

	ViewCount    *int `json:"viewCount,omitempty"`
	LikeCount    *int `json:"likeCount,omitempty"`
	RetweetCount *int `json:"retweetCount,omitempty"`
	ReplyCount   *int `json:"replyCount,omitempty"`
	CommentCount *int `json:"commentCount,omitempty"`
	QuoteCount   *int `json:"quoteCount,omitempty"`
	DanmakuCount *int `json:"danmakuCount,omitempty"`

	PublishedAt *time.Time `json:"publishedAt,omitempty"`
}

type AnalysisResult struct {
	IsReal           bool       `json:"isReal"`
	Relevance        int        `json:"relevance"`
	RelevanceReason  string     `json:"relevanceReason"`
	KeywordMentioned bool       `json:"keywordMentioned"`
	Importance       Importance `json:"importance"`
	Summary          string     `json:"summary"`
}

type Status struct {
	Total    int            `json:"total"`
	Today    int            `json:"today"`
	Urgent   int            `json:"urgent"`
	BySource map[string]int `json:"bySource"`
}
