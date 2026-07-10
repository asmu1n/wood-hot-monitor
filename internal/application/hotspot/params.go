package hotspot

import (
	"strings"
	"time"

	domain "wood-hot-monitor/internal/domain/hotspot"
)

type SortBy string

const (
	SortByCreatedAt   SortBy = "createdAt"
	SortByRelevance   SortBy = "relevance"
	SortByImportance  SortBy = "importance"
	SortByPublishedAt SortBy = "publishedAt"
	SortByLikeCount   SortBy = "likeCount"
	SortByViewCount   SortBy = "viewCount"
)

type GetAllParams struct {
	Page       int                  `json:"page"`
	Limit      int                  `json:"limit"`
	Source     *string              `json:"source"`
	Importance *domain.Importance   `json:"importance"`
	KeywordID  *string              `json:"keywordId"`
	IsReal     *bool                `json:"isReal"`
	TimeRange  *string              `json:"timeRange"`
	TimeFrom   *string              `json:"timeFrom"`
	TimeTo     *string              `json:"timeTo"`
	SortBy     *SortBy              `json:"sortBy"`
	SortOrder  *string              `json:"sortOrder"`
}

type SearchParams struct {
	Page    int      `json:"page"`
	Limit   int      `json:"limit"`
	Query   string   `json:"query"`
	Sources []string `json:"sources"`
}

var sortByMapping = map[SortBy]domain.SortField{
	"createdAt":   domain.SortByCreatedAt,
	"relevance":   domain.SortByRelevance,
	"importance":  domain.SortByImportance,
	"publishedAt": domain.SortByPublishedAt,
	"likeCount":   domain.SortByLikeCount,
	"viewCount":   domain.SortByViewCount,
}

func (p *GetAllParams) ToFilter() domain.Filter {
	f := domain.Filter{
		Source:     p.Source,
		Importance: p.Importance,
		KeywordID:  p.KeywordID,
		IsReal:     p.IsReal,
		SortBy:     domain.SortByCreatedAt,
		SortOrder:  domain.SortDesc,
		Page:       p.Page,
		Limit:      p.Limit,
	}

	if p.SortBy != nil {
		if mapped, ok := sortByMapping[*p.SortBy]; ok {
			f.SortBy = mapped
		}
	}
	if p.SortOrder != nil && strings.ToUpper(*p.SortOrder) == "ASC" {
		f.SortOrder = domain.SortAsc
	}

	timeFrom, timeTo := resolveTimeRange(p.TimeRange, p.TimeFrom, p.TimeTo)
	f.TimeFrom = timeFrom
	f.TimeTo = timeTo

	return f
}

func resolveTimeRange(timeRange, timeFrom, timeTo *string) (*time.Time, *time.Time) {
	if timeFrom != nil && timeTo != nil {
		from, _ := time.Parse(time.DateTime, *timeFrom)
		to, _ := time.Parse(time.DateTime, *timeTo)
		return &from, &to
	}
	if timeRange == nil {
		return nil, nil
	}
	now := time.Now().UTC()
	switch *timeRange {
	case "1h":
		t := now.Add(-time.Hour)
		return &t, nil
	case "24h":
		t := now.Add(-24 * time.Hour)
		return &t, nil
	case "7d":
		t := now.AddDate(0, 0, -7)
		return &t, nil
	case "30d":
		t := now.AddDate(0, 0, -30)
		return &t, nil
	default:
		return nil, nil
	}
}
