package checker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"wood-hot-monitor/internal/config"
	"wood-hot-monitor/internal/hotspot"
	"wood-hot-monitor/internal/keyword"
	"wood-hot-monitor/internal/infra/notify"
	"wood-hot-monitor/internal/infra/scraper"
)

const (
	freshnessWindow     = 7 * 24 * time.Hour
	maxResultsPerSource = 5
	minRelevance        = 30
)

type Service struct {
	keyword  *keyword.Service
	cfg      *config.Service
	analyzer hotspot.Analyzer
	hotspot  *hotspot.Service
	scraper  hotspot.Scraper
	notifier notify.Notifier
}

func NewService(
	kw *keyword.Service,
	cfg *config.Service,
	analyzer hotspot.Analyzer,
	hs *hotspot.Service,
	scraperSvc hotspot.Scraper,
	notifier notify.Notifier,
) *Service {
	return &Service{
		keyword:  kw,
		cfg:      cfg,
		analyzer: analyzer,
		hotspot:  hs,
		scraper:  scraperSvc,
		notifier: notifier,
	}
}

func (s *Service) Run(ctx context.Context) error {
	s.notifier.Emit(notify.EventCheckerStarted, nil)
	defer s.notifier.Emit(notify.EventCheckerCompleted, nil)

	keywords, err := s.keyword.GetAll(ctx, true)
	if err != nil {
		return fmt.Errorf("list keywords: %w", err)
	}

	if len(keywords) == 0 {
		log.Println("checker: no active keywords, skipping")
		return nil
	}

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

			expanded, err := s.analyzer.ExpandKeyword(ctx, kw.Text)
			if err != nil {
				log.Printf("checker: expand keyword %q failed: %v", kw.Text, err)
				expanded = []string{kw.Text}
			}

			allResults := s.scraper.SearchAll(ctx, kw.Text, hotspot.ScraperConfig{
				TwitterAPIKey: twitterAPIKey,
			})
			allResults = scraper.DeduplicateByURL(allResults)
			allResults = scraper.FilterByFreshness(allResults, freshnessWindow)
			scraper.SortByPriority(allResults)
			scored := scraper.FilterAndSort(allResults)
			limited := applySourceQuota(scored, maxResultsPerSource)

			var contents []string
			for _, sr := range limited {
				contents = append(contents, sr.Title+" "+sr.Content)
			}

			analysisResults, err := s.analyzer.BatchAnalyze(ctx, contents, kw.Text, expanded)
			if err != nil {
				log.Printf("checker: batch analyze failed: %v", err)
				return
			}

			for i, ar := range analysisResults {
				if ar.Relevance < minRelevance {
					continue
				}

				targetSearchResult := limited[i].SearchResult

				hotspotID, isNew, err := s.hotspot.UpsertFromSearch(ctx, targetSearchResult, ar, &kw.ID)
				if err != nil {
					log.Printf("checker: upsert hotspot failed: %v", err)
					continue
				}

				if isNew {
					s.notifier.Emit(notify.EventHotspotNew, map[string]string{
						"id":     hotspotID,
						"title":  targetSearchResult.Title,
						"source": targetSearchResult.Source,
					})
					log.Printf("checker: new hotspot: %s", targetSearchResult.Title)
					if ar.Importance == hotspot.ImportanceHigh || ar.Importance == hotspot.ImportanceUrgent {
						notify.SendEmailAlert(cfg, targetSearchResult, ar)
					}
				}
			}
		}()
	}

	wg.Wait()
	return nil
}

func applySourceQuota(results []scraper.ScoredResult, maxPerSource int) []scraper.ScoredResult {
	counts := make(map[string]int)
	var out []scraper.ScoredResult
	for _, r := range results {
		if counts[r.Source] >= maxPerSource {
			continue
		}
		counts[r.Source]++
		out = append(out, r)
	}
	return out
}
