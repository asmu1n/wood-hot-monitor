package checker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	"wood-hot-monitor/ent"
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

			hotspotID, isNew, err := s.upsertHotspot(ctx, sr.SearchResult, analysis, &kw.ID)
			if err != nil {
				log.Printf("checker: upsert hotspot failed: %v", err)
				continue
			}

			if isNew {
				s.emit("hotspot:new", map[string]string{
					"id":     hotspotID,
					"title":  sr.Title,
					"source": sr.Source,
				})

				if analysis.Importance == "high" || analysis.Importance == "urgent" {
					s.sendEmailAlert(cfg, sr.SearchResult, analysis)
				}
			}
		}
	}

	log.Printf("checker: run complete, %d AI analyses used", aiCount)
	return nil
}

func (s *Service) upsertHotspot(ctx context.Context, r scraper.SearchResult, analysis *llm.AnalysisResult, keywordID *string) (string, bool, error) {
	now := time.Now().UTC()

	existing, err := s.db.Client.Hotspot.Query().
		Where(enthot.URLEQ(r.URL), enthot.SourceEQ(r.Source)).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return "", false, err
	}

	if existing != nil {
		builder := s.db.Client.Hotspot.UpdateOneID(existing.ID).
			SetTitle(r.Title).
			SetContent(r.Content).
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
			SetNillablePublishedAt(r.PublishedAt)

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

		if err := builder.Exec(ctx); err != nil {
			return "", false, err
		}
		return existing.ID, false, nil
	}

	builder := s.db.Client.Hotspot.Create().
		SetID(uuid.NewString()).
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
		SetIsNotified(true).
		SetNotifiedAt(now).
		SetIsRead(false).
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

	h, err := builder.Save(ctx)
	if err != nil {
		return "", false, err
	}
	return h.ID, true, nil
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
