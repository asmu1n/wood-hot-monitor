package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"wood-hot-monitor/internal/domain/hotspot"
)

const twitterAPIURL = "https://api.twitterapi.io/twitter/tweet/advanced_search"

type twitterResponse struct {
	Tweets     []twitterTweet `json:"tweets"`
	HasNext    bool           `json:"has_next"`
	NextCursor string         `json:"next_cursor"`
}

type twitterTweet struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	CreatedAt string `json:"createdAt"`
	URL       string `json:"url"`
	Author    struct {
		Name           string `json:"name"`
		UserName       string `json:"userName"`
		ProfilePicture string `json:"profilePicture"`
		Followers      int    `json:"followers"`
		IsBlueVerified bool   `json:"isBlueVerified"`
	} `json:"author"`
	LikeCount    int `json:"likeCount"`
	RetweetCount int `json:"retweetCount"`
	ReplyCount   int `json:"replyCount"`
	QuoteCount   int `json:"quoteCount"`
	ViewCount    int `json:"viewCount"`
}

func SearchTwitter(ctx context.Context, query, apiKey string) ([]hotspot.SearchResult, error) {
	if apiKey == "" {
		return nil, nil
	}

	params := url.Values{
		"query":     {query},
		"queryType": {"Latest"},
	}
	reqURL := fmt.Sprintf("%s?%s", twitterAPIURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Accept", "application/json")

	client := NewHTTPClient(20 * time.Second)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("twitter request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("twitter status %d", resp.StatusCode)
	}

	var data twitterResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode twitter response: %w", err)
	}

	results := make([]hotspot.SearchResult, 0, len(data.Tweets))
	for _, tweet := range data.Tweets {
		tweetURL := tweet.URL
		if tweetURL == "" {
			tweetURL = fmt.Sprintf("https://x.com/%s/status/%s", tweet.Author.UserName, tweet.ID)
		}

		var published *time.Time
		if tweet.CreatedAt != "" {
			if t, err := time.Parse("Mon Jan 02 15:04:05 +0000 2006", tweet.CreatedAt); err == nil {
				published = TimePtr(t)
			}
		}

		results = append(results, hotspot.SearchResult{
			Title:    truncate(tweet.Text, 100),
			Content:  tweet.Text,
			URL:      tweetURL,
			Source:   "twitter",
			SourceID: tweet.ID,
			Author: &hotspot.Author{
				Name:      tweet.Author.Name,
				Username:  tweet.Author.UserName,
				Avatar:    tweet.Author.ProfilePicture,
				Followers: tweet.Author.Followers,
				Verified:  tweet.Author.IsBlueVerified,
			},
			ViewCount:    IntPtr(tweet.ViewCount),
			LikeCount:    IntPtr(tweet.LikeCount),
			RetweetCount: IntPtr(tweet.RetweetCount),
			ReplyCount:   IntPtr(tweet.ReplyCount),
			QuoteCount:   IntPtr(tweet.QuoteCount),
			PublishedAt:  published,
		})
	}

	return results, nil
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return s
}
