package hotspot

import "time"

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
	ViewCount        *int
	LikeCount        *int
	RetweetCount     *int
	ReplyCount       *int
	CommentCount     *int
	QuoteCount       *int
	DanmakuCount     *int
	Author           *Author
	PublishedAt      *time.Time
	CreatedAt        time.Time
	KeywordID        *string
	Keyword          *HotspotKeyword
	IsNotified       bool
	NotifiedAt       *time.Time
	IsRead           bool
}

type HotspotKeyword struct {
	ID       string
	Text     string
	Category *string
}

type Author struct {
	Name      string
	Username  string
	Avatar    string
	Followers int
	Verified  bool
}

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

type AnalysisResult struct {
	IsReal           bool       `json:"isReal"`
	Relevance        int        `json:"relevance"`
	RelevanceReason  string     `json:"relevanceReason"`
	KeywordMentioned bool       `json:"keywordMentioned"`
	Importance       Importance `json:"importance"`
	Summary          string     `json:"summary"`
}

type Status struct {
	Total    int
	Today    int
	Urgent   int
	BySource map[string]int
}
