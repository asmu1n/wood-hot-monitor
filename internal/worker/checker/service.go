package checker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	"wood-hot-monitor/internal/biz/hotspot"
	"wood-hot-monitor/internal/biz/keyword"
	"wood-hot-monitor/internal/biz/quality"
	"wood-hot-monitor/internal/biz/scraper"

	"wood-hot-monitor/internal/core/config"
	"wood-hot-monitor/internal/core/event"
	"wood-hot-monitor/internal/infra/email"
	"wood-hot-monitor/internal/infra/llm"
)

const (
	freshnessWindow     = 7 * 24 * time.Hour
	maxResultsPerSource = 5
	minRelevance        = 30
)

type Service struct {
	keyword  *keyword.Service
	cfg      *config.Service
	llm      *llm.Service
	hotspot  *hotspot.Service
	notifier event.Notifier
}

func NewService(keyword *keyword.Service, cfg *config.Service, llm *llm.Service, hotspot *hotspot.Service, notifier event.Notifier) *Service {
	return &Service{keyword, cfg, llm, hotspot, notifier}
}

func (s *Service) Run(ctx context.Context) error {
	// 事件通知
	s.notifier.Emit(event.EventCheckerStarted, nil)
	defer s.notifier.Emit(event.EventCheckerCompleted, nil)

	// 获取活跃关键词
	keywords, err := s.keyword.GetAll(true)
	if err != nil {
		return fmt.Errorf("list keywords: %w", err)
	}

	if len(keywords) == 0 {
		log.Println("checker: no active keywords, skipping")
		return nil
	}

	// 读取本地配置
	cfg, err := s.cfg.Get()
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	twitterAPIKey, _ := cfg.Settings["twitterApiKey"].(string)

	var wg sync.WaitGroup
	for _, kw := range keywords {
		if ctx.Err() != nil {
			break
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			expanded, err := s.llm.ExpandKeyword(ctx, kw.Text)
			if err != nil {
				log.Printf("checker: expand keyword %q failed: %v", kw.Text, err)
				expanded = []string{kw.Text}
			}

			allResults := s.searchAllSources(ctx, kw.Text, twitterAPIKey)
			allResults = scraper.DeduplicateByURL(allResults)
			allResults = scraper.FilterByFreshness(allResults, freshnessWindow)
			scraper.SortByPriority(allResults)
			scored := quality.FilterAndSort(allResults)
			limited := applySourceQuota(scored, maxResultsPerSource)

			var contents []string
			for _, sr := range limited {
				contents = append(contents, sr.Title+" "+sr.Content)
			}

			analysisResults, err := s.llm.BatchAnalyze(ctx, contents, kw.Text, expanded)
			if err != nil {
				log.Printf("checker: batch analyze failed: %v", err)
				return
			}

			for i, ar := range analysisResults {
				if ar.Relevance < minRelevance {
					continue
				}

				targetSearchResult := limited[i].SearchResult

				hotspotID, isNew, err := s.hotspot.UpsertHotspot(ctx, targetSearchResult, ar, &kw.ID)
				if err != nil {
					log.Printf("checker: upsert hotspot failed: %v", err)
					continue
				}

				if isNew {
					s.notifier.Emit(event.EventHotspotNew, map[string]string{
						"id":     hotspotID,
						"title":  targetSearchResult.Title,
						"source": targetSearchResult.Source,
					})

					if ar.Importance == "high" || ar.Importance == "urgent" {
						email.SendEmailAlert(cfg, targetSearchResult, ar)
					}
				}
			}
		}()
	}

	wg.Wait()

	return nil
}

// 爬取所有来源
func (s *Service) searchAllSources(ctx context.Context, query, twitterAPIKey string) []scraper.SearchResult {
	type sourceResult struct {
		results []scraper.SearchResult
		source  string
		err     error
	}

	ch := make(chan sourceResult, 4)
	var wg sync.WaitGroup

	search := func(name string, fn func() ([]scraper.SearchResult, error)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results, err := fn()
			ch <- sourceResult{results: results, source: name, err: err}
		}()
	}

	search("hackernews", func() ([]scraper.SearchResult, error) {
		return scraper.SearchHackerNews(ctx, query)
	})
	search("bing", func() ([]scraper.SearchResult, error) {
		return scraper.SearchBing(ctx, query)
	})
	search("bilibili", func() ([]scraper.SearchResult, error) {
		return scraper.SearchBilibili(ctx, query)
	})
	search("twitter", func() ([]scraper.SearchResult, error) {
		return scraper.SearchTwitter(ctx, query, twitterAPIKey)
	})

	go func() {
		wg.Wait()
		close(ch)
	}()

	var all []scraper.SearchResult
	for sr := range ch {
		if sr.err != nil {
			log.Printf("checker: %s search failed: %v", sr.source, sr.err)
			continue
		}
		all = append(all, sr.results...)
	}

	return all
}

// 数据数量过滤
func applySourceQuota(results []quality.ScoredResult, maxPerSource int) []quality.ScoredResult {
	counts := make(map[string]int)
	var out []quality.ScoredResult
	for _, r := range results {
		if counts[r.Source] >= maxPerSource {
			continue
		}
		counts[r.Source]++
		out = append(out, r)
	}
	return out
}
