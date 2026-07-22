package keyword

import "time"

type Keyword struct {
	ID           string    `json:"id"`
	Text         string    `json:"text"`
	Category     *string   `json:"category"`
	IsActive     bool      `json:"isActive"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	HotspotCount *int      `json:"hotspotCount"`
}
