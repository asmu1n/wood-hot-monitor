package hotspot

type SortField string

const (
	SortByCreatedAt   SortField = "createdAt"
	SortByRelevance   SortField = "relevance"
	SortByImportance  SortField = "importance"
	SortByPublishedAt SortField = "publishedAt"
	SortByLikeCount   SortField = "likeCount"
	SortByViewCount   SortField = "viewCount"
)

type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)
