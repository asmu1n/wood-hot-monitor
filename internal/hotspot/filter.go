package hotspot

import "time"

type Filter struct {
	Source     *string
	Importance *Importance
	KeywordID  *string
	IsReal     *bool
	TimeFrom   *time.Time
	TimeTo     *time.Time
	SortBy     SortField
	SortOrder  SortOrder
	Page       int
	Limit      int
}

func (f *Filter) Offset() int {
	return (f.Page - 1) * f.Limit
}

type SearchFilter struct {
	Query   string
	Sources []string
	Page    int
	Limit   int
}

func (f *SearchFilter) Offset() int {
	return (f.Page - 1) * f.Limit
}
