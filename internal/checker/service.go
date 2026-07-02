package checker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	enthot "wood-hot-monitor/ent/hotspot"
	entkw "wood-hot-monitor/ent/keyword"
	"wood-hot-monitor/internal/config"
	"wood-hot-monitor/internal/database"
	"wood-hot-monitor/internal/email"
	"wood-hot-monitor/internal/llm"
	"wood-hot-monitor/internal/models"
	"wood-hot-monitor/internal/quality"
	"wood-hot-monitor/internal/scraper"
)

const (
	freshnessWindow     = 7 * 24 * time.Hour
	maxResultsPerSource = 5
	globalAICap         = 12
	minRelevance        = 30
)

type EventEmitter func(eventName string, data any)

type Service struct {
	db      *database.DB
	cfg     *config.Service
	llm     *llm.Service
	emitter EventEmitter
}

func NewService(db *database.DB, cfg *config.Service, llmSvc *llm.Service, emitter EventEmitter) *Service {
	return &Service{db: db, cfg: cfg, llm: llmSvc, emitter: emitter}
}

func (s *Service) Run(ctx context.Context) error {
	s.emit("check:start", nil)
	defer s.emit("check:complete", nil)

	keywords, err := s.db.Client.Keyword.Query().
		Where(entkw.IsActive(true)).
		All(ctx)
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
	aiCount := 0

	for _, kw := range keywords {
		if ctx.Err() != nil {
			return ctx.Err()
		}

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

		for _, sr := range limited {
			preMatch := llm.PreMatchKeyword(sr.Title+" "+sr.Content, expanded)
			analysis, err := s.llm.AnalyzeContent(ctx, sr.Content, kw.Text, &preMatch)
			if err != nil {
				log.Printf("checker: analyze content failed: %v", err)
				continue
			}
			aiCount++

			if analysis.Relevance < minRelevance {
				continue
			}

			hotspotID := uuid.NewString()
			if err := s.upsertHotspot(ctx, hotspotID, sr.SearchResult, analysis, &kw.ID); err != nil {
				log.Printf("checker: upsert hotspot failed: %v", err)
				continue
			}

			s.createNotification(ctx, hotspotID, sr.SearchResult, analysis)

			if analysis.Importance == "high" || analysis.Importance == "urgent" {
				s.sendEmailAlert(cfg, sr.SearchResult, analysis)
			}

			s.emit("hotspot:new", map[string]string{
				"id":     hotspotID,
				"title":  sr.Title,
				"source": sr.Source,
			})
		}
	}

	log.Printf("checker: run complete, %d AI analyses used", aiCount)
	return nil
}

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

func (s *Service) upsertHotspot(ctx context.Context, id string, r scraper.SearchResult, analysis *llm.AnalysisResult, keywordID *string) error {
	now := time.Now().UTC()

	builder := s.db.Client.Hotspot.Create().
		SetID(id).
		SetTitle(r.Title).
		SetContent(r.Content).
		SetURL(r.URL).
		SetSource(r.Source).
		SetNillableSourceID(strPtr(r.SourceID)).
		SetIsReal(analysis.IsReal).
		SetRelevance(analysis.Relevance).
		SetNillableRelevanceReason(strPtr(analysis.RelevanceReason)).
		SetNillableKeywordMentioned(&analysis.KeywordMentioned).
		SetImportance(analysis.Importance).
		SetNillableSummary(strPtr(analysis.Summary)).
		SetNillableViewCount(r.ViewCount).
		SetNillableLikeCount(r.LikeCount).
		SetNillableRetweetCount(r.RetweetCount).
		SetNillableReplyCount(r.ReplyCount).
		SetNillableCommentCount(r.CommentCount).
		SetNillableQuoteCount(r.QuoteCount).
		SetNillableDanmakuCount(r.DanmakuCount).
		SetNillablePublishedAt(r.PublishedAt).
		SetCreatedAt(now)

	if r.Author != nil {
		builder = builder.
			SetNillableAuthorName(strPtr(r.Author.Name)).
			SetNillableAuthorUsername(strPtr(r.Author.Username)).
			SetNillableAuthorAvatar(strPtr(r.Author.Avatar))
		if r.Author.Followers > 0 {
			builder = builder.SetAuthorFollowers(r.Author.Followers)
		}
		if r.Author.Verified {
			builder = builder.SetAuthorVerified(true)
		}
	}

	if keywordID != nil {
		builder = builder.SetKeywordID(*keywordID)
	}

	return builder.
		OnConflictColumns(enthot.FieldURL, enthot.FieldSource).
		UpdateNewValues().
		Exec(ctx)
}

func (s *Service) createNotification(ctx context.Context, hotspotID string, r scraper.SearchResult, analysis *llm.AnalysisResult) {
	title := fmt.Sprintf("[%s] %s", r.Source, r.Title)
	content := analysis.Summary
	if content == "" {
		content = r.Content
		if len([]rune(content)) > 100 {
			content = string([]rune(content)[:100]) + "..."
		}
	}

	notifType := "info"
	if analysis.Importance == "urgent" {
		notifType = "urgent"
	} else if analysis.Importance == "high" {
		notifType = "warning"
	}

	nBuilder := s.db.Client.Notification.Create().
		SetType(notifType).
		SetTitle(title).
		SetContent(content).
		SetIsRead(false).
		SetCreatedAt(time.Now().UTC()).
		SetHotspotID(hotspotID)

	row, err := nBuilder.Save(ctx)
	if err != nil {
		log.Printf("checker: create notification failed: %v", err)
		return
	}

	s.emit("notification:new", map[string]string{"id": row.ID, "type": notifType, "title": title})
}

func (s *Service) sendEmailAlert(cfg *models.AppConfig, r scraper.SearchResult, analysis *llm.AnalysisResult) {
	resendKey, _ := cfg.Settings["resendApiKey"].(string)
	if resendKey == "" || cfg.EmailAddress == "" {
		return
	}

	emailSvc := email.NewService(resendKey, "noreply@woodmonitor.app")
	summary := analysis.Summary
	if summary == "" {
		summary = r.Content
	}
	if err := emailSvc.SendHotspotAlert(cfg.EmailAddress, r.Title, r.Source, analysis.Importance, summary, r.URL); err != nil {
		log.Printf("checker: send email failed: %v", err)
	}
}

func (s *Service) emit(eventName string, data any) {
	if s.emitter != nil {
		s.emitter(eventName, data)
	}
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
