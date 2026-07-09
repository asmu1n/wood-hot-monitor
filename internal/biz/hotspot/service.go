package hotspot

import (
	"context"
	"strings"
	"time"

	"wood-hot-monitor/ent"
	hsmodel "wood-hot-monitor/ent/hotspot"
	"wood-hot-monitor/ent/predicate"
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

type GetAllParams struct {
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	Source     *string            `json:"source"`
	Importance *models.Importance `json:"importance"`
	KeywordID  *string            `json:"keywordId"`
	IsReal     *bool              `json:"isReal"`
	TimeRange  *string            `json:"timeRange"`
	TimeFrom   *string            `json:"timeFrom"`
	TimeTo     *string            `json:"timeTo"`
	SortBy     *string            `json:"sortBy"`
	SortOrder  *string            `json:"sortOrder"`
}

type SearchParams struct {
	Page    int      `json:"page"`
	Limit   int      `json:"limit"`
	Query   string   `json:"query"`
	Sources []string `json:"sources"`
}

var allowedSortFields = map[string]string{
	"createdAt":   hsmodel.FieldCreatedAt,
	"relevance":   hsmodel.FieldRelevance,
	"importance":  hsmodel.FieldImportance,
	"publishedAt": hsmodel.FieldPublishedAt,
	"likeCount":   hsmodel.FieldLikeCount,
	"viewCount":   hsmodel.FieldViewCount,
}

func (s *Service) GetAll(ctx context.Context, params GetAllParams) (*models.PaginatedResult[models.Hotspot], error) {
	var preds []predicate.Hotspot

	timeFrom, timeTo := resolveTimeRange(params.TimeRange, params.TimeFrom, params.TimeTo)

	preds = append(preds, optionalVal(params.Source, hsmodel.SourceEQ)...)
	preds = append(preds, optionalVal(params.Importance, hsmodel.ImportanceEQ)...)
	preds = append(preds, optionalVal(params.KeywordID, hsmodel.KeywordIDEQ)...)
	preds = append(preds, optionalVal(params.IsReal, hsmodel.IsRealEQ)...)
	preds = append(preds, optionalVal(timeFrom, hsmodel.CreatedAtGTE)...)
	preds = append(preds, optionalVal(timeTo, hsmodel.CreatedAtLTE)...)

	query := s.client.Hotspot.Query()
	if len(preds) > 0 {
		query = query.Where(preds...)
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}

	orderField := hsmodel.FieldCreatedAt
	if params.SortBy != nil {
		if f, ok := allowedSortFields[*params.SortBy]; ok {
			orderField = f
		}
	}
	orderFn := ent.Desc(orderField)
	if params.SortOrder != nil && strings.ToUpper(*params.SortOrder) == "ASC" {
		orderFn = ent.Asc(orderField)
	}

	offset := (params.Page - 1) * params.Limit
	rows, err := query.
		Order(orderFn).
		Limit(params.Limit).
		Offset(offset).
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
	var preds []predicate.Hotspot
	preds = append(preds, hsmodel.Or(
		hsmodel.TitleContains(params.Query),
		hsmodel.ContentContains(params.Query),
	))
	if len(params.Sources) > 0 {
		preds = append(preds, hsmodel.SourceIn(params.Sources...))
	}

	query := s.client.Hotspot.Query().Where(preds...)

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}

	offset := (params.Page - 1) * params.Limit

	rows, err := query.
		Order(ent.Desc(hsmodel.FieldRelevance), ent.Desc(hsmodel.FieldCreatedAt)).
		Limit(params.Limit).
		Offset(offset).
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

func mapHotspot(h *ent.Hotspot) models.Hotspot {
	m := models.Hotspot{
		ID:               h.ID,
		Title:            h.Title,
		Content:          h.Content,
		URL:              h.URL,
		Source:           h.Source,
		SourceID:         h.SourceID,
		IsReal:           h.IsReal,
		Relevance:        h.Relevance,
		RelevanceReason:  h.RelevanceReason,
		KeywordMentioned: h.KeywordMentioned,
		Importance:       h.Importance,
		Summary:          h.Summary,
		ViewCount:        h.ViewCount,
		LikeCount:        h.LikeCount,
		RetweetCount:     h.RetweetCount,
		ReplyCount:       h.ReplyCount,
		CommentCount:     h.CommentCount,
		QuoteCount:       h.QuoteCount,
		DanmakuCount:     h.DanmakuCount,
		AuthorName:       h.AuthorName,
		AuthorUsername:   h.AuthorUsername,
		AuthorAvatar:     h.AuthorAvatar,
		AuthorFollowers:  h.AuthorFollowers,
		AuthorVerified:   h.AuthorVerified,
		KeywordID:        h.KeywordID,
		IsNotified:       h.IsNotified,
		IsRead:           h.IsRead,
		CreatedAt:        h.CreatedAt.Format(time.DateTime),
	}
	if h.NotifiedAt != nil {
		s := h.NotifiedAt.Format(time.DateTime)
		m.NotifiedAt = &s
	}
	if h.PublishedAt != nil {
		s := h.PublishedAt.Format(time.DateTime)
		m.PublishedAt = &s
	}
	if h.Edges.Keyword != nil {
		kw := h.Edges.Keyword
		m.Keyword = &models.HotspotKeyword{ID: kw.ID, Text: kw.Text, Category: kw.Category}
	}
	return m
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func optionalVal[T any, P any](val *T, fn func(T) P) []P {
	if val == nil {
		return nil
	}
	return []P{fn(*val)}
}
