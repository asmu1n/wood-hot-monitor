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

const bilibiliSearchURL = "https://api.bilibili.com/x/web-interface/search/all/v2"

type bilibiliResponse struct {
	Code int `json:"code"`
	Data struct {
		Result []bilibiliResultGroup `json:"result"`
	} `json:"data"`
}

type bilibiliResultGroup struct {
	ResultType string           `json:"result_type"`
	Data       []bilibiliResult `json:"data"`
}

type bilibiliResult struct {
	BVid        string `json:"bvid"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Mid         int    `json:"mid"`
	Pic         string `json:"pic"`
	Play        int    `json:"play"`
	Danmaku     int    `json:"danmaku"`
	Like        int    `json:"like"`
	Reply       int    `json:"reply"`
	Favorites   int    `json:"favorites"`
	Description string `json:"description"`
	PubDate     int64  `json:"pubdate"`
	ArcURL      string `json:"arcurl"`
}

func SearchBilibili(ctx context.Context, query string) ([]hotspot.SearchResult, error) {
	params := url.Values{
		"keyword":     {query},
		"search_type": {"video"},
	}
	reqURL := fmt.Sprintf("%s?%s", bilibiliSearchURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	SetRequestHeaders(req)
	req.Header.Set("Referer", "https://www.bilibili.com")

	client := NewHTTPClient(15 * time.Second)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bilibili request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bilibili status %d", resp.StatusCode)
	}

	var data bilibiliResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode bilibili response: %w", err)
	}

	if data.Code != 0 {
		return nil, fmt.Errorf("bilibili api error code %d", data.Code)
	}

	var results []hotspot.SearchResult
	for _, group := range data.Data.Result {
		if group.ResultType != "video" {
			continue
		}
		for _, item := range group.Data {
			itemURL := item.ArcURL
			if itemURL == "" {
				itemURL = fmt.Sprintf("https://www.bilibili.com/video/%s", item.BVid)
			}

			var published *time.Time
			if item.PubDate > 0 {
				t := time.Unix(item.PubDate, 0)
				published = &t
			}

			results = append(results, hotspot.SearchResult{
				Title:    stripHTMLTags(item.Title),
				Content:  item.Description,
				URL:      itemURL,
				Source:   "bilibili",
				SourceID: item.BVid,
				Author: &hotspot.Author{
					Name:     item.Author,
					Username: fmt.Sprintf("%d", item.Mid),
					Avatar:   item.Pic,
				},
				ViewCount:    IntPtr(item.Play),
				LikeCount:    IntPtr(item.Like),
				CommentCount: IntPtr(item.Reply),
				DanmakuCount: IntPtr(item.Danmaku),
				PublishedAt:  published,
			})
		}
	}

	return results, nil
}

func stripHTMLTags(s string) string {
	var result []byte
	inTag := false
	for i := 0; i < len(s); i++ {
		if s[i] == '<' {
			inTag = true
			continue
		}
		if s[i] == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result = append(result, s[i])
		}
	}
	return string(result)
}
