package ent

import (
	"context"
	"time"

	"wood-hot-monitor/ent"
	hsmodel "wood-hot-monitor/ent/hotspot"
	"wood-hot-monitor/ent/predicate"
	domain "wood-hot-monitor/internal/domain/hotspot"

	"github.com/google/uuid"
)

type HotspotRepository struct {
	client *ent.Client
}

func NewHotspotRepository(client *ent.Client) *HotspotRepository {
	return &HotspotRepository{client: client}
}

func (r *HotspotRepository) FindAll(ctx context.Context, filter domain.Filter) ([]domain.Hotspot, int, error) {
	preds := r.buildPredicates(filter)

	query := r.client.Hotspot.Query()
	if len(preds) > 0 {
		query = query.Where(preds...)
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	orderFunc := r.buildOrder(filter)
	rows, err := query.
		Order(orderFunc).
		Limit(filter.Limit).
		Offset(filter.Offset()).
		WithKeyword().
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]domain.Hotspot, len(rows))
	for i, row := range rows {
		result[i] = mapToDomain(row)
	}
	return result, total, nil
}

func (r *HotspotRepository) FindByID(ctx context.Context, id string) (*domain.Hotspot, error) {
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
	h := mapToDomain(row)
	return &h, nil
}

func (r *HotspotRepository) Search(ctx context.Context, filter domain.SearchFilter) ([]domain.Hotspot, int, error) {
	preds := []predicate.Hotspot{
		hsmodel.Or(
			hsmodel.TitleContains(filter.Query),
			hsmodel.ContentContains(filter.Query),
		),
	}
	if len(filter.Sources) > 0 {
		preds = append(preds, hsmodel.SourceIn(filter.Sources...))
	}

	query := r.client.Hotspot.Query().Where(preds...)

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	rows, err := query.
		Order(ent.Desc(hsmodel.FieldRelevance), ent.Desc(hsmodel.FieldCreatedAt)).
		Limit(filter.Limit).
		Offset(filter.Offset()).
		WithKeyword().
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]domain.Hotspot, len(rows))
	for i, row := range rows {
		result[i] = mapToDomain(row)
	}
	return result, total, nil
}

func (r *HotspotRepository) Upsert(ctx context.Context, h domain.Hotspot) (string, bool, error) {
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

func (r *HotspotRepository) GetStatus(ctx context.Context) (*domain.Status, error) {
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

	return &domain.Status{
		Total:    total,
		Today:    today,
		Urgent:   urgent,
		BySource: bySource,
	}, nil
}

func (r *HotspotRepository) GetNotifications(ctx context.Context, limit int) ([]domain.Hotspot, error) {
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

	result := make([]domain.Hotspot, len(rows))
	for i, row := range rows {
		result[i] = mapToDomain(row)
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

var sortFieldMap = map[domain.SortField]string{
	domain.SortByCreatedAt:   hsmodel.FieldCreatedAt,
	domain.SortByRelevance:   hsmodel.FieldRelevance,
	domain.SortByImportance:  hsmodel.FieldImportance,
	domain.SortByPublishedAt: hsmodel.FieldPublishedAt,
	domain.SortByLikeCount:   hsmodel.FieldLikeCount,
	domain.SortByViewCount:   hsmodel.FieldViewCount,
}

func (r *HotspotRepository) buildOrder(filter domain.Filter) hsmodel.OrderOption {
	field := hsmodel.FieldCreatedAt
	if f, ok := sortFieldMap[filter.SortBy]; ok {
		field = f
	}
	if filter.SortOrder == domain.SortAsc {
		return ent.Asc(field)
	}
	return ent.Desc(field)
}

func (r *HotspotRepository) buildPredicates(filter domain.Filter) []predicate.Hotspot {
	var preds []predicate.Hotspot

	if filter.Source != nil {
		preds = append(preds, hsmodel.SourceEQ(*filter.Source))
	}
	if filter.Importance != nil {
		preds = append(preds, hsmodel.ImportanceEQ(*filter.Importance))
	}
	if filter.KeywordID != nil {
		preds = append(preds, hsmodel.KeywordIDEQ(*filter.KeywordID))
	}
	if filter.IsReal != nil {
		preds = append(preds, hsmodel.IsRealEQ(*filter.IsReal))
	}
	if filter.TimeFrom != nil {
		preds = append(preds, hsmodel.CreatedAtGTE(*filter.TimeFrom))
	}
	if filter.TimeTo != nil {
		preds = append(preds, hsmodel.CreatedAtLTE(*filter.TimeTo))
	}

	return preds
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
