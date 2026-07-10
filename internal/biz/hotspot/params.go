package hotspot

import (
	"strings"
	"wood-hot-monitor/ent"
	hsmodel "wood-hot-monitor/ent/hotspot"
	"wood-hot-monitor/ent/predicate"
	"wood-hot-monitor/internal/core/models"
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
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	Source     *string            `json:"source"`
	Importance *models.Importance `json:"importance"`
	KeywordID  *string            `json:"keywordId"`
	IsReal     *bool              `json:"isReal"`
	TimeRange  *string            `json:"timeRange"`
	TimeFrom   *string            `json:"timeFrom"`
	TimeTo     *string            `json:"timeTo"`
	SortBy     *SortBy            `json:"sortBy"`
	SortOrder  *string            `json:"sortOrder"`
}

var allowedSortFields = map[SortBy]string{
	"createdAt":   hsmodel.FieldCreatedAt,
	"relevance":   hsmodel.FieldRelevance,
	"importance":  hsmodel.FieldImportance,
	"publishedAt": hsmodel.FieldPublishedAt,
	"likeCount":   hsmodel.FieldLikeCount,
	"viewCount":   hsmodel.FieldViewCount,
}

type SearchParams struct {
	Page    int      `json:"page"`
	Limit   int      `json:"limit"`
	Query   string   `json:"query"`
	Sources []string `json:"sources"`
}

func (p *GetAllParams) Predicates() []predicate.Hotspot {

	preds := make([]predicate.Hotspot, 6)

	timeFrom, timeTo := resolveTimeRange(p.TimeRange, p.TimeFrom, p.TimeTo)

	if p.Source != nil {
		preds = append(preds, hsmodel.SourceEQ(*p.Source))
	}

	if p.Importance != nil {
		preds = append(preds, hsmodel.ImportanceEQ(*p.Importance))
	}

	if p.KeywordID != nil {
		preds = append(preds, hsmodel.KeywordIDEQ(*p.KeywordID))
	}

	if p.IsReal != nil {
		preds = append(preds, hsmodel.IsRealEQ(*p.IsReal))
	}

	if timeFrom != nil {
		preds = append(preds, hsmodel.CreatedAtGTE(*timeFrom))
	}

	if timeTo != nil {
		preds = append(preds, hsmodel.CreatedAtLTE(*timeTo))
	}

	return preds

}

func (p *GetAllParams) OrderBy() hsmodel.OrderOption {
	// 排序字段
	orderField := hsmodel.FieldCreatedAt
	if p.SortBy != nil {
		if f, ok := allowedSortFields[*p.SortBy]; ok {
			orderField = f
		}
	}
	// 排序方式
	orderFn := ent.Desc(orderField)
	if p.SortOrder != nil && strings.ToUpper(*p.SortOrder) == "ASC" {
		orderFn = ent.Asc(orderField)
	}
	return orderFn
}

func (p *GetAllParams) Offset() int {
	return (p.Page - 1) * p.Limit
}

func (p *SearchParams) Predicates() []predicate.Hotspot {
	preds := make([]predicate.Hotspot, 3)
	preds = append(preds, hsmodel.Or(
		hsmodel.TitleContains(p.Query),
		hsmodel.ContentContains(p.Query),
	))
	if len(p.Sources) > 0 {
		preds = append(preds, hsmodel.SourceIn(p.Sources...))
	}
	return preds
}

func (p *SearchParams) OrderBy() []hsmodel.OrderOption {
	return []hsmodel.OrderOption{
		ent.Desc(hsmodel.FieldRelevance),
		ent.Desc(hsmodel.FieldCreatedAt),
	}
}

func (p *SearchParams) Offset() int {
	return (p.Page - 1) * p.Limit
}
