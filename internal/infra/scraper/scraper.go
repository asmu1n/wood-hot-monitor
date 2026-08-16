package scraper

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"wood-hot-monitor/internal/module/hotspot"
)

var userAgents = [4]string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:133.0) Gecko/20100101 Firefox/133.0",
}

func RandomUA() string {
	return userAgents[rand.Intn(len(userAgents))]
}

func NewHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}

// 模拟浏览器请求头
func SetRequestHeaders(req *http.Request) {
	req.Header.Set("User-Agent", RandomUA())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
}

type Scraper struct{}

func New() *Scraper {
	return &Scraper{}
}

func (s *Scraper) SearchAll(ctx context.Context, query string, config hotspot.ScraperConfig) []hotspot.SearchResult {
	type sourceResult struct {
		results []hotspot.SearchResult
		source  string
		err     error
	}

	type searchTask struct {
		name string
		fn   func() ([]hotspot.SearchResult, error)
	}

	// 定义搜索任务
	tasks := []searchTask{
		{"hackernews", func() ([]hotspot.SearchResult, error) {
			return SearchHackerNews(ctx, query)
		}},
		{"bing", func() ([]hotspot.SearchResult, error) {
			return SearchBing(ctx, query)
		}},
		{"bilibili", func() ([]hotspot.SearchResult, error) {
			return SearchBilibili(ctx, query)
		}},
		{"twitter", func() ([]hotspot.SearchResult, error) {
			return SearchTwitter(ctx, query, config.TwitterAPIKey)
		}},
	}

	// 创建通道和等待组
	ch := make(chan sourceResult, len(tasks))
	var wg sync.WaitGroup

	// 启动所有搜索任务
	for _, t := range tasks {
		wg.Add(1)

		go func(name string, fn func() ([]hotspot.SearchResult, error)) {
			defer wg.Done()

			// 错误捕获兜底
			defer func() {
				if r := recover(); r != nil {
					ch <- sourceResult{
						source: name,
						err:    fmt.Errorf("panic recovered: %v", r),
					}
				}
			}()
			results, err := fn()
			ch <- sourceResult{
				results: results,
				source:  name,
				err:     err,
			}
		}(t.name, t.fn)
	}

	// 等待任务完成然后关闭信道，让接收端处理全部结果
	go func() {
		wg.Wait()
		close(ch)
	}()

	all := make([]hotspot.SearchResult, 0, 20)

	// for select 持续尝试接收任务结果，并且在ctx取消时返回已收集的结果
	for {
		select {
		case sr, ok := <-ch:
			if !ok {
				return all
			}
			if sr.err != nil {
				log.Printf("scraper: %s search failed: %v", sr.source, sr.err)
				continue
			}
			all = append(all, sr.results...)

		case <-ctx.Done():
			return all
		}
	}

}
