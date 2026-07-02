package hotspot

import (
	"context"
	"strings"
	"time"

	"wood-hot-monitor/ent"
	enthot "wood-hot-monitor/ent/hotspot"
	"wood-hot-monitor/ent/predicate"
	"wood-hot-monitor/internal/database"
	"wood-hot-monitor/internal/models"
)

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

type GetAllParams struct {
	Page       int     `json:"page"`
	Limit      int     `json:"limit"`
	Source     *string `json:"source"`
	Importance *string `json:"importance"`
	KeywordID  *string `json:"keywordId"`
	IsReal     *bool   `json:"isReal"`
	TimeRange  *string `json:"timeRange"`
	TimeFrom   *string `json:"timeFrom"`
	TimeTo     *string `json:"timeTo"`
	SortBy     *string `json:"sortBy"`
	SortOrder  *string `json:"sortOrder"`
}

var allowedSortFields = map[string]string{
	"createdAt":   enthot.FieldCreatedAt,
	"relevance":   enthot.FieldRelevance,
	"importance":  enthot.FieldImportance,
	"publishedAt": enthot.FieldPublishedAt,
	"likeCount":   enthot.FieldLikeCount,
	"viewCount":   enthot.FieldViewCount,
}

func (s *Service) GetAll(params GetAllParams) (*models.PaginatedResult[models.Hotspot], error) {
	ctx := context.Background()
	var preds []predicate.Hotspot

	timeFrom, timeTo := resolveTimeRange(params.TimeRange, params.TimeFrom, params.TimeTo)

	preds = append(preds, database.OptionalVal(params.Source, enthot.SourceEQ)...)
	preds = append(preds, database.OptionalVal(params.Importance, enthot.ImportanceEQ)...)
	preds = append(preds, database.OptionalVal(params.KeywordID, enthot.KeywordIDEQ)...)
	preds = append(preds, database.OptionalVal(params.IsReal, enthot.IsRealEQ)...)
	preds = append(preds, database.OptionalVal(timeFrom, enthot.CreatedAtGTE)...)
	preds = append(preds, database.OptionalVal(timeTo, enthot.CreatedAtLTE)...)

	query := s.db.Client.Hotspot.Query()
	if len(preds) > 0 {
		query = query.Where(preds...)
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}

	orderField := enthot.FieldCreatedAt
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

func (s *Service) GetByID(id string) (*models.Hotspot, error) {
	row, err := s.db.Client.Hotspot.Query().
		Where(enthot.IDEQ(id)).
		WithKeyword().
		Only(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	m := mapHotspot(row)
	return &m, nil
}

func (s *Service) GetStatus() (*models.Status, error) {
	ctx := context.Background()

	total, _ := s.db.Client.Hotspot.Query().Count(ctx)

	todayStart := time.Now().UTC().Truncate(24 * time.Hour)
	today, _ := s.db.Client.Hotspot.Query().
		Where(enthot.CreatedAtGTE(todayStart)).Count(ctx)

	urgent, _ := s.db.Client.Hotspot.Query().
		Where(enthot.ImportanceEQ("urgent")).Count(ctx)

	bySource := make(map[string]int)
	var sourceCounts []struct {
		Source string `json:"source"`
		Count  int    `json:"count"`
	}
	err := s.db.Client.Hotspot.Query().
		GroupBy(enthot.FieldSource).
		Aggregate(ent.Count()).
		Scan(ctx, &sourceCounts)
	if err == nil {
		for _, sc := range sourceCounts {
			bySource[sc.Source] = sc.Count
		}
	}

	return &models.Status{
		Total:    total,
		Today:    today,
		Urgent:   urgent,
		BySource: bySource,
	}, nil
}

func (s *Service) Delete(id string) error {
	return s.db.Client.Hotspot.DeleteOneID(id).Exec(context.Background())
}

func (s *Service) Search(query string, sources []string) ([]models.Hotspot, error) {
	ctx := context.Background()

	var preds []predicate.Hotspot
	preds = append(preds, enthot.Or(
		enthot.TitleContains(query),
		enthot.ContentContains(query),
	))
	if len(sources) > 0 {
		preds = append(preds, enthot.SourceIn(sources...))
	}

	rows, err := s.db.Client.Hotspot.Query().
		Where(preds...).
		Order(ent.Desc(enthot.FieldRelevance), ent.Desc(enthot.FieldCreatedAt)).
		Limit(50).
		WithKeyword().
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]models.Hotspot, len(rows))
	for i, row := range rows {
		result[i] = mapHotspot(row)
	}
	return result, nil
}

func (s *Service) Check() error {
	return nil
}

func (s *Service) GetNotifications(limit int) ([]models.Hotspot, error) {
	ctx := context.Background()
	if limit <= 0 {
		limit = 10
	}

	rows, err := s.db.Client.Hotspot.Query().
		Where(enthot.IsNotified(true)).
		Order(ent.Desc(enthot.FieldNotifiedAt)).
		Limit(limit).
		WithKeyword().
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]models.Hotspot, len(rows))
	for i, row := range rows {
		result[i] = mapHotspot(row)
	}
	return result, nil
}

func (s *Service) UnreadCount() (int, error) {
	return s.db.Client.Hotspot.Query().
		Where(enthot.IsNotified(true), enthot.IsRead(false)).
		Count(context.Background())
}

func (s *Service) MarkRead(id string) error {
	return s.db.Client.Hotspot.UpdateOneID(id).
		SetIsRead(true).
		Exec(context.Background())
}

func (s *Service) MarkAllRead() error {
	return s.db.Client.Hotspot.Update().
		Where(enthot.IsNotified(true), enthot.IsRead(false)).
		SetIsRead(true).
		Exec(context.Background())
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
