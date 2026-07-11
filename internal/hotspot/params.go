package hotspot

import (
	"strings"
	"time"
)

type ParamsSortBy string

const (
	ParamsSortByCreatedAt   ParamsSortBy = "createdAt"
	ParamsSortByRelevance   ParamsSortBy = "relevance"
	ParamsSortByImportance  ParamsSortBy = "importance"
	ParamsSortByPublishedAt ParamsSortBy = "publishedAt"
	ParamsSortByLikeCount   ParamsSortBy = "likeCount"
	ParamsSortByViewCount   ParamsSortBy = "viewCount"
)

type GetAllParams struct {
	Page       int          `json:"page"`
	Limit      int          `json:"limit"`
	Source     *string      `json:"source"`
	Importance *Importance  `json:"importance"`
	KeywordID  *string      `json:"keywordId"`
	IsReal     *bool        `json:"isReal"`
	TimeRange  *string      `json:"timeRange"`
	TimeFrom   *string      `json:"timeFrom"`
	TimeTo     *string      `json:"timeTo"`
	SortBy     *ParamsSortBy `json:"sortBy"`
	SortOrder  *string      `json:"sortOrder"`
}

type SearchParams struct {
	Page    int      `json:"page"`
	Limit   int      `json:"limit"`
	Query   string   `json:"query"`
	Sources []string `json:"sources"`
}

var sortByMapping = map[ParamsSortBy]SortField{
	"createdAt":   SortByCreatedAt,
	"relevance":   SortByRelevance,
	"importance":  SortByImportance,
	"publishedAt": SortByPublishedAt,
	"likeCount":   SortByLikeCount,
	"viewCount":   SortByViewCount,
}

func (p *GetAllParams) ToFilter() Filter {
	f := Filter{
		Source:     p.Source,
		Importance: p.Importance,
		KeywordID:  p.KeywordID,
		IsReal:     p.IsReal,
		SortBy:     SortByCreatedAt,
		SortOrder:  SortDesc,
		Page:       p.Page,
		Limit:      p.Limit,
	}

	if p.SortBy != nil {
		if mapped, ok := sortByMapping[*p.SortBy]; ok {
			f.SortBy = mapped
		}
	}
	if p.SortOrder != nil && strings.ToUpper(*p.SortOrder) == "ASC" {
		f.SortOrder = SortAsc
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
