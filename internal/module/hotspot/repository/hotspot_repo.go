package repository

import (
	"context"
	"time"

	"wood-hot-monitor/ent"
	hsmodel "wood-hot-monitor/ent/hotspot"
	"wood-hot-monitor/ent/predicate"
	"wood-hot-monitor/internal/module/hotspot"

	"github.com/google/uuid"
)

type HotspotRepository struct {
	client *ent.Client
}

func NewHotspotRepository(client *ent.Client) *HotspotRepository {
	return &HotspotRepository{client: client}
}

func (r *HotspotRepository) FindAll(ctx context.Context, params hotspot.GetAllParams) ([]hotspot.Hotspot, int, error) {
	preds := r.buildPredicates(params)

	query := r.client.Hotspot.Query()
	if len(preds) > 0 {
		query = query.Where(preds...)
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	orderFunc := r.buildOrder(params)
	rows, err := query.
		Order(orderFunc).
		Limit(params.Limit()).
		Offset(params.Offset()).
		WithKeyword().
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]hotspot.Hotspot, len(rows))
	for i, row := range rows {
		result[i] = mapTomodule(row)
	}
	return result, total, nil
}

func (r *HotspotRepository) FindByID(ctx context.Context, id string) (*hotspot.Hotspot, error) {
	row, err := r.client.Hotspot.Query().
		Where(hsmodel.IDEQ(id)).
		WithKeyword().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	h := mapTomodule(row)
	return &h, nil
}

func (r *HotspotRepository) Search(ctx context.Context, params hotspot.SearchParams) ([]hotspot.Hotspot, int, error) {
	preds := []predicate.Hotspot{
		hsmodel.Or(
			hsmodel.TitleContains(params.Query),
			hsmodel.ContentContains(params.Query),
		),
	}
	if len(params.Sources) > 0 {
		preds = append(preds, hsmodel.SourceIn(params.Sources...))
	}

	query := r.client.Hotspot.Query().Where(preds...)

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	rows, err := query.
		Order(ent.Desc(hsmodel.FieldRelevance), ent.Desc(hsmodel.FieldCreatedAt)).
		Limit(params.Limit()).
		Offset(params.Offset()).
		WithKeyword().
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]hotspot.Hotspot, len(rows))
	for i, row := range rows {
		result[i] = mapTomodule(row)
	}
	return result, total, nil
}

func (r *HotspotRepository) Upsert(ctx context.Context, h hotspot.Hotspot) (string, bool, error) {
	now := time.Now().UTC()

	existing, err := r.client.Hotspot.Query().
		Where(hsmodel.URLEQ(h.URL), hsmodel.SourceEQ(h.Source)).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return "", false, err
	}

	if existing != nil {
		builder := r.client.Hotspot.UpdateOneID(existing.ID).
			SetTitle(h.Title).
			SetContent(h.Content).
			SetNillableSourceID(h.SourceID).
			SetIsReal(h.IsReal).
			SetRelevance(h.Relevance).
			SetNillableRelevanceReason(h.RelevanceReason).
			SetNillableKeywordMentioned(h.KeywordMentioned).
			SetImportance(h.Importance).
			SetNillableSummary(h.Summary).
			SetNillableViewCount(h.ViewCount).
			SetNillableLikeCount(h.LikeCount).
			SetNillableRetweetCount(h.RetweetCount).
			SetNillableReplyCount(h.ReplyCount).
			SetNillableCommentCount(h.CommentCount).
			SetNillableQuoteCount(h.QuoteCount).
			SetNillableDanmakuCount(h.DanmakuCount).
			SetNillablePublishedAt(h.PublishedAt)

		if h.Author != nil {
			builder = builder.
				SetNillableAuthorName(strPtr(h.Author.Name)).
				SetNillableAuthorUsername(strPtr(h.Author.Username)).
				SetNillableAuthorAvatar(strPtr(h.Author.Avatar))
			if h.Author.Followers > 0 {
				builder = builder.SetAuthorFollowers(h.Author.Followers)
			}
			if h.Author.Verified {
				builder = builder.SetAuthorVerified(true)
			}
		}

		if h.KeywordID != nil {
			builder = builder.SetKeywordID(*h.KeywordID)
		}

		if err := builder.Exec(ctx); err != nil {
			return "", false, err
		}
		return existing.ID, false, nil
	}

	id := h.ID
	if id == "" {
		id = uuid.NewString()
	}

	builder := r.client.Hotspot.Create().
		SetID(id).
		SetTitle(h.Title).
		SetContent(h.Content).
		SetURL(h.URL).
		SetSource(h.Source).
		SetNillableSourceID(h.SourceID).
		SetIsReal(h.IsReal).
		SetRelevance(h.Relevance).
		SetNillableRelevanceReason(h.RelevanceReason).
		SetNillableKeywordMentioned(h.KeywordMentioned).
		SetImportance(h.Importance).
		SetNillableSummary(h.Summary).
		SetNillableViewCount(h.ViewCount).
		SetNillableLikeCount(h.LikeCount).
		SetNillableRetweetCount(h.RetweetCount).
		SetNillableReplyCount(h.ReplyCount).
		SetNillableCommentCount(h.CommentCount).
		SetNillableQuoteCount(h.QuoteCount).
		SetNillableDanmakuCount(h.DanmakuCount).
		SetNillablePublishedAt(h.PublishedAt).
		SetIsNotified(h.IsNotified).
		SetIsRead(h.IsRead).
		SetCreatedAt(now)

	if h.NotifiedAt != nil {
		builder = builder.SetNotifiedAt(*h.NotifiedAt)
	}

	if h.Author != nil {
		builder = builder.
			SetNillableAuthorName(strPtr(h.Author.Name)).
			SetNillableAuthorUsername(strPtr(h.Author.Username)).
			SetNillableAuthorAvatar(strPtr(h.Author.Avatar))
		if h.Author.Followers > 0 {
			builder = builder.SetAuthorFollowers(h.Author.Followers)
		}
		if h.Author.Verified {
			builder = builder.SetAuthorVerified(true)
		}
	}

	if h.KeywordID != nil {
		builder = builder.SetKeywordID(*h.KeywordID)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		return "", false, err
	}
	return row.ID, true, nil
}

func (r *HotspotRepository) Delete(ctx context.Context, id string) error {
	return r.client.Hotspot.DeleteOneID(id).Exec(ctx)
}

func (r *HotspotRepository) GetStatus(ctx context.Context) (*hotspot.Status, error) {
	total, _ := r.client.Hotspot.Query().Count(ctx)

	todayStart := time.Now().UTC().Truncate(24 * time.Hour)
	today, _ := r.client.Hotspot.Query().
		Where(hsmodel.CreatedAtGTE(todayStart)).Count(ctx)

	urgent, _ := r.client.Hotspot.Query().
		Where(hsmodel.ImportanceEQ("urgent")).Count(ctx)

	bySource := make(map[string]int)
	var sourceCounts []struct {
		Source string `json:"source"`
		Count  int    `json:"count"`
	}
	err := r.client.Hotspot.Query().
		GroupBy(hsmodel.FieldSource).
		Aggregate(ent.Count()).
		Scan(ctx, &sourceCounts)
	if err == nil {
		for _, sc := range sourceCounts {
			bySource[sc.Source] = sc.Count
		}
	}

	return &hotspot.Status{
		Total:    total,
		Today:    today,
		Urgent:   urgent,
		BySource: bySource,
	}, nil
}

func (r *HotspotRepository) GetNotifications(ctx context.Context, limit int) ([]hotspot.Hotspot, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := r.client.Hotspot.Query().
		Where(hsmodel.IsNotified(true)).
		Order(ent.Desc(hsmodel.FieldNotifiedAt)).
		Limit(limit).
		WithKeyword().
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]hotspot.Hotspot, len(rows))
	for i, row := range rows {
		result[i] = mapTomodule(row)
	}
	return result, nil
}

func (r *HotspotRepository) UnreadCount(ctx context.Context) (int, error) {
	return r.client.Hotspot.Query().
		Where(hsmodel.IsNotified(true), hsmodel.IsRead(false)).
		Count(ctx)
}

func (r *HotspotRepository) MarkRead(ctx context.Context, id string) error {
	return r.client.Hotspot.UpdateOneID(id).
		SetIsRead(true).
		Exec(ctx)
}

func (r *HotspotRepository) MarkAllRead(ctx context.Context) error {
	return r.client.Hotspot.Update().
		Where(hsmodel.IsNotified(true), hsmodel.IsRead(false)).
		SetIsRead(true).
		Exec(ctx)
}

var sortFieldColumn = map[hotspot.SortField]string{
	hotspot.SortByCreatedAt:   hsmodel.FieldCreatedAt,
	hotspot.SortByRelevance:   hsmodel.FieldRelevance,
	hotspot.SortByImportance:  hsmodel.FieldImportance,
	hotspot.SortByPublishedAt: hsmodel.FieldPublishedAt,
	hotspot.SortByLikeCount:   hsmodel.FieldLikeCount,
	hotspot.SortByViewCount:   hsmodel.FieldViewCount,
}

func (r *HotspotRepository) buildOrder(params hotspot.GetAllParams) hsmodel.OrderOption {
	field := hsmodel.FieldCreatedAt
	if f, ok := sortFieldColumn[params.SortBy]; ok {
		field = f
	}
	if params.SortOrder == hotspot.SortAsc {
		return ent.Asc(field)
	}
	return ent.Desc(field)
}

func (r *HotspotRepository) buildPredicates(params hotspot.GetAllParams) []predicate.Hotspot {
	var preds []predicate.Hotspot

	if params.Source != nil {
		preds = append(preds, hsmodel.SourceEQ(*params.Source))
	}
	if params.Importance != nil {
		preds = append(preds, hsmodel.ImportanceEQ(*params.Importance))
	}
	if params.KeywordID != nil {
		preds = append(preds, hsmodel.KeywordIDEQ(*params.KeywordID))
	}
	if params.IsReal != nil {
		preds = append(preds, hsmodel.IsRealEQ(*params.IsReal))
	}
	if params.TimeFrom != nil {
		preds = append(preds, hsmodel.CreatedAtGTE(*params.TimeFrom))
	}
	if params.TimeTo != nil {
		preds = append(preds, hsmodel.CreatedAtLTE(*params.TimeTo))
	}

	return preds
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func mapTomodule(row *ent.Hotspot) hotspot.Hotspot {
	h := hotspot.Hotspot{
		ID:               row.ID,
		Title:            row.Title,
		Content:          row.Content,
		URL:              row.URL,
		Source:           row.Source,
		SourceID:         row.SourceID,
		IsReal:           row.IsReal,
		Relevance:        row.Relevance,
		RelevanceReason:  row.RelevanceReason,
		KeywordMentioned: row.KeywordMentioned,
		Importance:       row.Importance,
		Summary:          row.Summary,
		ViewCount:        row.ViewCount,
		LikeCount:        row.LikeCount,
		RetweetCount:     row.RetweetCount,
		ReplyCount:       row.ReplyCount,
		CommentCount:     row.CommentCount,
		QuoteCount:       row.QuoteCount,
		DanmakuCount:     row.DanmakuCount,
		KeywordID:        row.KeywordID,
		IsNotified:       row.IsNotified,
		IsRead:           row.IsRead,
		CreatedAt:        row.CreatedAt,
		PublishedAt:      row.PublishedAt,
		NotifiedAt:       row.NotifiedAt,
	}

	if row.AuthorName != nil || row.AuthorUsername != nil {
		h.Author = &hotspot.Author{
			Name:     deref(row.AuthorName),
			Username: deref(row.AuthorUsername),
			Avatar:   deref(row.AuthorAvatar),
		}
		if row.AuthorFollowers != nil {
			h.Author.Followers = *row.AuthorFollowers
		}
		if row.AuthorVerified != nil {
			h.Author.Verified = *row.AuthorVerified
		}
	}

	if row.Edges.Keyword != nil {
		kw := row.Edges.Keyword
		h.Keyword = &hotspot.HotspotKeyword{
			ID:       kw.ID,
			Text:     kw.Text,
			Category: kw.Category,
		}
	}
	return h
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
