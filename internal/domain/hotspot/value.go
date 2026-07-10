package hotspot

// Importance 热点重要度值对象
type Importance string

const (
	ImportanceLow    Importance = "low"
	ImportanceMedium Importance = "medium"
	ImportanceHigh   Importance = "high"
	ImportanceUrgent Importance = "urgent"
)

func (i Importance) Values() []string {
	return []string{
		string(ImportanceLow),
		string(ImportanceMedium),
		string(ImportanceHigh),
		string(ImportanceUrgent),
	}
}

func (i Importance) IsValid() bool {
	switch i {
	case ImportanceLow, ImportanceMedium, ImportanceHigh, ImportanceUrgent:
		return true
	}
	return false
}

// SortField 排序字段
type SortField string

const (
	SortByCreatedAt   SortField = "created_at"
	SortByRelevance   SortField = "relevance"
	SortByImportance  SortField = "importance"
	SortByPublishedAt SortField = "published_at"
	SortByLikeCount   SortField = "like_count"
	SortByViewCount   SortField = "view_count"
)

// SortOrder 排序方向
type SortOrder string

const (
	SortAsc  SortOrder = "ASC"
	SortDesc SortOrder = "DESC"
)
