package hotspot

import "time"

type GetAllParams struct {
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	Source     *string     `json:"source"`
	Importance *Importance `json:"importance"`
	KeywordID  *string     `json:"keywordId"`
	IsReal     *bool       `json:"isReal"`
	TimeRange  *string     `json:"timeRange"`
	TimeFrom   *time.Time  `json:"timeFrom"`
	TimeTo     *time.Time  `json:"timeTo"`
	SortBy     SortField   `json:"sortBy"`
	SortOrder  SortOrder   `json:"sortOrder"`
}

func (p *GetAllParams) Offset() int { return (p.Page - 1) * p.Limit }

func (p *GetAllParams) Resolve() {
	if p.SortBy == "" {
		p.SortBy = SortByCreatedAt
	}
	if p.SortOrder == "" {
		p.SortOrder = SortDesc
	}
	if p.TimeRange != nil && p.TimeFrom == nil {
		now := time.Now().UTC()
		switch *p.TimeRange {
		case "1h":
			t := now.Add(-time.Hour)
			p.TimeFrom = &t
		case "24h":
			t := now.Add(-24 * time.Hour)
			p.TimeFrom = &t
		case "7d":
			t := now.AddDate(0, 0, -7)
			p.TimeFrom = &t
		case "30d":
			t := now.AddDate(0, 0, -30)
			p.TimeFrom = &t
		}
	}
}

type SearchParams struct {
	Page    int      `json:"page"`
	Limit   int      `json:"limit"`
	Query   string   `json:"query"`
	Sources []string `json:"sources"`
}

func (p *SearchParams) Offset() int { return (p.Page - 1) * p.Limit }
