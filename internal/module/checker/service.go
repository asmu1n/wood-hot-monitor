package checker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"wood-hot-monitor/internal/config"
	"wood-hot-monitor/internal/module/hotspot"
	"wood-hot-monitor/internal/module/keyword"
	"wood-hot-monitor/internal/port"
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
	notifier port.Notifier
}

func NewService(
	kw *keyword.Service,
	cfg *config.Service,
	analyzer hotspot.Analyzer,
	hs *hotspot.Service,
	scraperSvc hotspot.Scraper,
	notifier port.Notifier,
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
	// emit 搜刮热点信息事件进度
	s.notifier.Emit(port.EventCheckerStarted, nil)
	defer s.notifier.Emit(port.EventCheckerCompleted, nil)

	// 获取当前活跃监听的关键词
	keywords, err := s.keyword.GetAll(ctx, true)
	if err != nil {
		return fmt.Errorf("list keywords: %w", err)
	}

	if len(keywords) == 0 {
		log.Println("checker: no active keywords, skipping")
		return nil
	}

	// 获取当前本地设置快照
	cfg, err := s.cfg.Get()
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	twitterAPIKey := cfg.TwitterApiKey

	// 初始化等待锁，准备并发任务
	var wg sync.WaitGroup
	for _, kw := range keywords {
		// 判断当前任务上下文状态，如果以及取消就主动退出
		if ctx.Err() != nil {
			break
		}

		// 协程任务计数（要在函数体外执行）
		wg.Add(1)
		go func() {
			// 当前协程任务完成标记
			defer wg.Done()

			// 扩展关键词（服务异常就回退到原始关键词）
			expanded, err := s.analyzer.ExpandKeyword(ctx, kw.Text)
			if err != nil {
				log.Printf("checker: expand keyword %q failed: %v", kw.Text, err)
				expanded = []string{kw.Text}
			}

			// 开始爬取结果
			allResults := s.scraper.SearchAll(ctx, kw.Text, hotspot.ScraperConfig{
				TwitterAPIKey: twitterAPIKey,
			})
			// 来源 URL 去重过滤
			allResults = deduplicateByURL(allResults)
			// 时间新鲜度过滤
			allResults = filterByFreshness(allResults, freshnessWindow)
			// 信息来源权重排序
			sortByPriority(allResults)
			// 分析热点信息质量并过滤低质量信息(评判标准是作者信息和互动数据)
			scored := filterAndSort(allResults)
			// 按来源配额限制再过滤一遍
			limited := applySourceQuota(scored, maxResultsPerSource)

			// 整合热点标题和内容，准备内容分析
			contents := make([]string, len(limited))
			for i, sr := range limited {
				contents[i] = sr.Title + " " + sr.Content
			}

			// 分析内容相关性
			analysisResults, err := s.analyzer.BatchAnalyze(ctx, contents, kw.Text, expanded)
			if err != nil {
				log.Printf("checker: batch analyze failed: %v", err)
				return
			}

			// 对分析结果进行筛选，相关性较高才录入数据库
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

				// 新热点：应用内事件 + 按 config 策略分发 OS 通知 / 邮件
				if isNew {
					log.Printf("checker: new hotspot: %s", targetSearchResult.Title)
					s.notifier.OnHotspotNew(cfg.NotifyConfig, port.HotspotAlert{
						ID:         hotspotID,
						Title:      targetSearchResult.Title,
						Source:     targetSearchResult.Source,
						URL:        targetSearchResult.URL,
						Importance: ar.Importance,
						Summary:    ar.Summary,
					})
				}
			}
		}()
	}

	// 阻塞等待所有goroutine完成
	wg.Wait()
	return nil
}

func applySourceQuota(results []ScoredResult, maxPerSource int) []ScoredResult {
	counts := make(map[string]int)
	var out []ScoredResult
	for _, r := range results {
		if counts[r.Source] >= maxPerSource {
			continue
		}
		counts[r.Source]++
		out = append(out, r)
	}
	return out
}
