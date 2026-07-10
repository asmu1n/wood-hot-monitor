package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	domain "wood-hot-monitor/internal/domain/hotspot"
)

const hnAlgoliaURL = "http://hn.algolia.com/api/v1/search"

type hnResponse struct {
	Hits []hnHit `json:"hits"`
}

type hnHit struct {
	ObjectID    string  `json:"objectID"`
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	Author      string  `json:"author"`
	Points      int     `json:"points"`
	NumComments int     `json:"num_comments"`
	CreatedAt   string  `json:"created_at"`
	StoryText   *string `json:"story_text"`
}

func SearchHackerNews(ctx context.Context, query string) ([]domain.SearchResult, error) {
	params := url.Values{
		"query": {query},
		"tags":  {"story"},
	}
	reqURL := fmt.Sprintf("%s?%s", hnAlgoliaURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	SetRequestHeaders(req)

	client := NewHTTPClient(15 * time.Second)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hn request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hn status %d", resp.StatusCode)
	}

	var data hnResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode hn response: %w", err)
	}

	results := make([]domain.SearchResult, 0, len(data.Hits))
	for _, hit := range data.Hits {
		itemURL := hit.URL
		if itemURL == "" {
			itemURL = fmt.Sprintf("https://news.ycombinator.com/item?id=%s", hit.ObjectID)
		}

		content := ""
		if hit.StoryText != nil {
			content = *hit.StoryText
		}

		var published *time.Time
		if t, err := time.Parse(time.RFC3339, hit.CreatedAt); err == nil {
			published = TimePtr(t)
		}

		results = append(results, domain.SearchResult{
			Title:    hit.Title,
			Content:  content,
			URL:      itemURL,
			Source:   "hackernews",
			SourceID: hit.ObjectID,
			Author: &domain.Author{
				Name:     hit.Author,
				Username: hit.Author,
			},
			LikeCount:    IntPtr(hit.Points),
			CommentCount: IntPtr(hit.NumComments),
			PublishedAt:  published,
		})
	}

	return results, nil
}
