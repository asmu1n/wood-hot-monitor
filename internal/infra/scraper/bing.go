package scraper

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"wood-hot-monitor/internal/module/hotspot"
)

const bingSearchURL = "https://www.bing.com/search"

func SearchBing(ctx context.Context, query string) ([]hotspot.SearchResult, error) {
	params := url.Values{"q": {query}}
	reqURL := fmt.Sprintf("%s?%s", bingSearchURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	SetRequestHeaders(req)

	client := NewHTTPClient(15 * time.Second)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bing status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse bing html: %w", err)
	}

	var results []hotspot.SearchResult
	doc.Find("li.b_algo").Each(func(i int, s *goquery.Selection) {
		titleEl := s.Find("h2 a")
		title := strings.TrimSpace(titleEl.Text())
		href, exists := titleEl.Attr("href")
		if !exists || title == "" {
			return
		}

		snippet := strings.TrimSpace(s.Find(".b_caption p").Text())
		if snippet == "" {
			snippet = strings.TrimSpace(s.Find(".b_caption .b_algoSlug").Text())
		}

		results = append(results, hotspot.SearchResult{
			Title:   title,
			Content: snippet,
			URL:     href,
			Source:  "bing",
		})
	})

	return results, nil
}
