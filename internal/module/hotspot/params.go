package hotspot

import (
	"time"
	"wood-hot-monitor/pkg/page"
	"wood-hot-monitor/pkg/types"
)

type GetAllParams struct {
	page.PageRequest
	Source     *string           `json:"source"`
	Importance *types.Importance `json:"importance"`
	KeywordID  *string           `json:"keywordId"`
	IsReal     *bool             `json:"isReal"`
	TimeRange  *string           `json:"timeRange"`
	TimeFrom   *time.Time        `json:"timeFrom"`
	TimeTo     *time.Time        `json:"timeTo"`
	SortBy     SortField         `json:"sortBy"`
	SortOrder  SortOrder         `json:"sortOrder"`
}

type SearchParams struct {
	page.PageRequest
	Query   string   `json:"query"`
	Sources []string `json:"sources"`
}

// DeleteParams 条件批量删除参数。
// PublishedAt / CreatedAt 语义为「早于等于该时间」的记录；
// MaxRelevance 语义为关联度 ≤ 该值；
// MaxImportance 语义为重要程度不高于该等级（low < medium < high < urgent）。
// 至少需要提供一个条件，否则拒绝执行以防误删全表。
type DeleteParams struct {
	KeywordID     *string           `json:"keywordId"`
	PublishedAt   *time.Time        `json:"publishedAt"`
	CreatedAt     *time.Time        `json:"createdAt"`
	IsRead        *bool             `json:"isRead"`
	MaxRelevance  *int              `json:"maxRelevance"`
	MaxImportance *types.Importance `json:"maxImportance"`
}

// HasFilter 是否至少包含一个有效删除条件。
func (p DeleteParams) HasFilter() bool {
	return p.KeywordID != nil ||
		p.PublishedAt != nil ||
		p.CreatedAt != nil ||
		p.IsRead != nil ||
		p.MaxRelevance != nil ||
		p.MaxImportance != nil
}
