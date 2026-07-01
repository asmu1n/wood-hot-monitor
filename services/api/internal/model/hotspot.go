package model

// Hotspot represents a monitored hotspot item.
type Hotspot struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Source   string `json:"source"`
	URL      string `json:"url"`
	Score    float64 `json:"score"`
	// TODO: add remaining fields matching current schema
}
