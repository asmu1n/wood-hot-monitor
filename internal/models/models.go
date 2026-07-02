package models

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
	Importance       string          `json:"importance"`
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
}

type HotspotKeyword struct {
	ID       string  `json:"id"`
	Text     string  `json:"text"`
	Category *string `json:"category"`
}

type Notification struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`
	Title     string  `json:"title"`
	Content   string  `json:"content"`
	IsRead    bool    `json:"isRead"`
	HotSpotID *string `json:"hotSpotId"`
	CreatedAt string  `json:"createdAt"`
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
	LLMProvider   string         `json:"llmProvider"`
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
