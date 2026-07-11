package keyword

import "time"

type Keyword struct {
	ID           string
	Text         string
	Category     *string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	HotspotCount *int
}
