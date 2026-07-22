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
