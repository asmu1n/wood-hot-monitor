package keyword

import (
	"context"
	"time"
	"wood-hot-monitor/ent"
	kwmodel "wood-hot-monitor/ent/keyword"
	"wood-hot-monitor/internal/core/models"
)

type Service struct {
	client *ent.Client
}

func NewService(client *ent.Client) *Service {
	return &Service{client: client}
}

func (s *Service) GetAll(ctx context.Context, activeOnly bool) ([]models.Keyword, error) {
	query := s.client.Keyword.Query().Order(ent.Desc(kwmodel.FieldCreatedAt))
	if activeOnly {
		query = query.Where(kwmodel.IsActive(true))
	}

	rows, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]models.Keyword, len(rows))
	for i, row := range rows {
		hCount, _ := row.QueryHotspots().Count(ctx)
		result[i] = mapKeyword(row, &hCount)
	}

	return result, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*models.Keyword, error) {
	row, err := s.client.Keyword.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	m := mapKeyword(row, nil)
	return &m, nil
}

func (s *Service) Create(ctx context.Context, text string, category *string) (*models.Keyword, error) {
	now := time.Now().UTC()
	builder := s.client.Keyword.Create().
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
	m := mapKeyword(row, nil)
	return &m, nil
}

func (s *Service) Update(ctx context.Context, id string, text *string, category *string) (*models.Keyword, error) {
	now := time.Now().UTC()
	builder := s.client.Keyword.UpdateOneID(id).SetUpdatedAt(now)
	if text != nil {
		builder = builder.SetText(*text)
	}
	if category != nil {
		builder = builder.SetCategory(*category)
	}

	if _, err := builder.Save(ctx); err != nil {
		return nil, err
	}
	return s.GetByID(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.client.Keyword.DeleteOneID(id).Exec(ctx)
}

func (s *Service) Toggle(ctx context.Context, id string) (*models.Keyword, error) {
	row, err := s.client.Keyword.Get(ctx, id)
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
	m := mapKeyword(row, nil)
	return &m, nil
}

func mapKeyword(row *ent.Keyword, hotspotCount *int) models.Keyword {
	return models.Keyword{
		ID:           row.ID,
		Text:         row.Text,
		Category:     row.Category,
		IsActive:     row.IsActive,
		CreatedAt:    row.CreatedAt.Format(time.DateTime),
		UpdatedAt:    row.UpdatedAt.Format(time.DateTime),
		HotspotCount: hotspotCount,
	}
}
