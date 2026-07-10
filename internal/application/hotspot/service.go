package hotspot

import (
	"context"
	"time"

	domain "wood-hot-monitor/internal/domain/hotspot"
	"wood-hot-monitor/internal/domain/shared"
)

type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetAll(ctx context.Context, params GetAllParams) (*shared.PaginatedResult[domain.Hotspot], error) {
	filter := params.ToFilter()
	data, total, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &shared.PaginatedResult[domain.Hotspot]{
		Data:  data,
		Total: total,
		Page:  params.Page,
		Limit: params.Limit,
	}, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*domain.Hotspot, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) Search(ctx context.Context, params SearchParams) (*shared.PaginatedResult[domain.Hotspot], error) {
	filter := domain.SearchFilter{
		Query:   params.Query,
		Sources: params.Sources,
		Page:    params.Page,
		Limit:   params.Limit,
	}
	data, total, err := s.repo.Search(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &shared.PaginatedResult[domain.Hotspot]{
		Data:  data,
		Total: total,
		Page:  params.Page,
		Limit: params.Limit,
	}, nil
}

func (s *Service) UpsertFromSearch(ctx context.Context, r domain.SearchResult, analysis *domain.AnalysisResult, keywordID *string) (string, bool, error) {
	now := time.Now().UTC()
	h := domain.Hotspot{
		Title:    r.Title,
		Content:  r.Content,
		URL:      r.URL,
		Source:   r.Source,
		SourceID: strPtr(r.SourceID),
		IsReal:   analysis.IsReal,
		Relevance: analysis.Relevance,
		RelevanceReason: strPtr(analysis.RelevanceReason),
		KeywordMentioned: &analysis.KeywordMentioned,
		Importance: analysis.Importance,
		Summary:    strPtr(analysis.Summary),
		ViewCount:    r.ViewCount,
		LikeCount:    r.LikeCount,
		RetweetCount: r.RetweetCount,
		ReplyCount:   r.ReplyCount,
		CommentCount: r.CommentCount,
		QuoteCount:   r.QuoteCount,
		DanmakuCount: r.DanmakuCount,
		PublishedAt:  r.PublishedAt,
		KeywordID:   keywordID,
		IsNotified:  true,
		NotifiedAt:  &now,
		IsRead:      false,
	}

	if r.Author != nil {
		h.Author = r.Author
	}

	return s.repo.Upsert(ctx, h)
}

func (s *Service) GetStatus(ctx context.Context) (*domain.Status, error) {
	return s.repo.GetStatus(ctx)
}

func (s *Service) GetNotifications(ctx context.Context, limit int) ([]domain.Hotspot, error) {
	return s.repo.GetNotifications(ctx, limit)
}

func (s *Service) UnreadCount(ctx context.Context) (int, error) {
	return s.repo.UnreadCount(ctx)
}

func (s *Service) MarkRead(ctx context.Context, id string) error {
	return s.repo.MarkRead(ctx, id)
}

func (s *Service) MarkAllRead(ctx context.Context) error {
	return s.repo.MarkAllRead(ctx)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
