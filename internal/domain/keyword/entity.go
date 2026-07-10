package keyword

import "time"

// Keyword 关键词聚合根
type Keyword struct {
	ID           string
	Text         string
	Category     *string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	HotspotCount *int
}
