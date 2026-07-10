package hotspot

import (
	"context"
	"time"

	"wood-hot-monitor/ent"
	hsmodel "wood-hot-monitor/ent/hotspot"
	"wood-hot-monitor/internal/core/models"

	"github.com/google/uuid"
)

type CheckFunc func(ctx context.Context) error

type Service struct {
	client *ent.Client
}

func NewService(client *ent.Client) *Service {
	return &Service{client: client}
}

func (s *Service) GetAll(ctx context.Context, params GetAllParams) (*models.PaginatedResult[models.Hotspot], error) {

	preds := params.Predicates()

	query := s.client.Hotspot.Query()
	if len(preds) > 0 {
		query = query.Where(preds...)
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := query.
		Order(params.OrderBy()).
		Limit(params.Limit).
		Offset(params.Offset()).
		WithKeyword().
		All(ctx)
	if err != nil {
		return nil, err
	}

	data := make([]models.Hotspot, len(rows))
	for i, row := range rows {
		data[i] = mapHotspot(row)
	}

	return &models.PaginatedResult[models.Hotspot]{
		Data:  data,
		Total: total,
		Page:  params.Page,
		Limit: params.Limit,
	}, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*models.Hotspot, error) {
	row, err := s.client.Hotspot.Query().
		Where(hsmodel.IDEQ(id)).
		WithKeyword().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	m := mapHotspot(row)
	return &m, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.client.Hotspot.DeleteOneID(id).Exec(ctx)
}

func (s *Service) Search(ctx context.Context, params SearchParams) (*models.PaginatedResult[models.Hotspot], error) {
	preds := params.Predicates()

	query := s.client.Hotspot.Query().Where(preds...)

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := query.
		Order(params.OrderBy()...).
		Limit(params.Limit).
		Offset(params.Offset()).
		WithKeyword().
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]models.Hotspot, len(rows))
	for i, row := range rows {
		result[i] = mapHotspot(row)
	}
	return &models.PaginatedResult[models.Hotspot]{
		Data:  result,
		Total: total,
		Page:  params.Page,
		Limit: params.Limit,
	}, nil
}

func (s *Service) UpsertHotspot(ctx context.Context, r models.SearchResult, analysis *models.AnalysisResult, keywordID *string) (string, bool, error) {
	now := time.Now().UTC()

	existing, err := s.client.Hotspot.Query().
		Where(hsmodel.URLEQ(r.URL), hsmodel.SourceEQ(r.Source)).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return "", false, err
	}

	if existing != nil {
		builder := s.client.Hotspot.UpdateOneID(existing.ID).
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

	builder := s.client.Hotspot.Create().
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

func resolveTimeRange(timeRange, timeFrom, timeTo *string) (*time.Time, *time.Time) {
	if timeFrom != nil && timeTo != nil {
		from, _ := time.Parse(time.DateTime, *timeFrom)
		to, _ := time.Parse(time.DateTime, *timeTo)
		return &from, &to
	}
	if timeRange == nil {
		return nil, nil
	}
	now := time.Now().UTC()
	switch *timeRange {
	case "1h":
		t := now.Add(-time.Hour)
		return &t, nil
	case "24h":
		t := now.Add(-24 * time.Hour)
		return &t, nil
	case "7d":
		t := now.AddDate(0, 0, -7)
		return &t, nil
	case "30d":
		t := now.AddDate(0, 0, -30)
		return &t, nil
	default:
		return nil, nil
	}
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
