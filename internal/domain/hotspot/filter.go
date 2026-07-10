package hotspot

import "time"

// Filter 热点查询过滤条件
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

// SearchFilter 搜索过滤条件
type SearchFilter struct {
	Query   string
	Sources []string
	Page    int
	Limit   int
}

func (f *SearchFilter) Offset() int {
	return (f.Page - 1) * f.Limit
}
