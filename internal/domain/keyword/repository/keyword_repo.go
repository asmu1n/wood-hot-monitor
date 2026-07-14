package repository

import (
	"context"
	"time"

	"wood-hot-monitor/ent"
	kwmodel "wood-hot-monitor/ent/keyword"
	"wood-hot-monitor/internal/domain/keyword"

	"github.com/google/uuid"
)

type KeywordRepository struct {
	client *ent.Client
}

func NewKeywordRepository(client *ent.Client) *KeywordRepository {
	return &KeywordRepository{client: client}
}

func (r *KeywordRepository) FindAll(ctx context.Context, activeOnly bool) ([]keyword.Keyword, error) {
	query := r.client.Keyword.Query().Order(ent.Desc(kwmodel.FieldCreatedAt))
	if activeOnly {
		query = query.Where(kwmodel.IsActive(true))
	}

	rows, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]keyword.Keyword, len(rows))
	for i, row := range rows {
		hCount, _ := row.QueryHotspots().Count(ctx)
		result[i] = mapKeywordToDomain(row, &hCount)
	}
	return result, nil
}

func (r *KeywordRepository) FindByID(ctx context.Context, id string) (*keyword.Keyword, error) {
	row, err := r.client.Keyword.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	kw := mapKeywordToDomain(row, nil)
	return &kw, nil
}

func (r *KeywordRepository) Create(ctx context.Context, text string, category *string) (*keyword.Keyword, error) {
	now := time.Now().UTC()
	builder := r.client.Keyword.Create().
		SetID(uuid.NewString()).
		SetText(text).
		SetIsActive(true).
		SetCreatedAt(now).
		SetUpdatedAt(now)
	if category != nil {
		builder = builder.SetCategory(*category)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	result := mapKeywordToDomain(row, nil)
	return &result, nil
}

func (r *KeywordRepository) Update(ctx context.Context, id string, text *string, category *string) (*keyword.Keyword, error) {
	now := time.Now().UTC()
	builder := r.client.Keyword.UpdateOneID(id).SetUpdatedAt(now)
	if text != nil {
		builder = builder.SetText(*text)
	}
	if category != nil {
		builder = builder.SetCategory(*category)
	}

	if _, err := builder.Save(ctx); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *KeywordRepository) Delete(ctx context.Context, id string) error {
	return r.client.Keyword.DeleteOneID(id).Exec(ctx)
}

func (r *KeywordRepository) Toggle(ctx context.Context, id string) (*keyword.Keyword, error) {
	row, err := r.client.Keyword.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	row, err = row.Update().
		SetIsActive(!row.IsActive).
		SetUpdatedAt(time.Now().UTC()).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	result := mapKeywordToDomain(row, nil)
	return &result, nil
}

func mapKeywordToDomain(row *ent.Keyword, hotspotCount *int) keyword.Keyword {
	return keyword.Keyword{
		ID:           row.ID,
		Text:         row.Text,
		Category:     row.Category,
		IsActive:     row.IsActive,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		HotspotCount: hotspotCount,
	}
}
