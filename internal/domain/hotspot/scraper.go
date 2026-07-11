package hotspot

import "context"

type Scraper interface {
	SearchAll(ctx context.Context, query string, config ScraperConfig) []SearchResult
}

type ScraperConfig struct {
	TwitterAPIKey string
}
