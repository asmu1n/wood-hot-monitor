package keyword

import (
	"context"
	"time"

	"wood-hot-monitor/ent"
	entkw "wood-hot-monitor/ent/keyword"
	"wood-hot-monitor/internal/database"
	"wood-hot-monitor/internal/models"
)

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

func (s *Service) GetAll(activeOnly bool) ([]models.Keyword, error) {
	ctx := context.Background()
	query := s.db.Client.Keyword.Query().Order(ent.Desc(entkw.FieldCreatedAt))
	if activeOnly {
		query = query.Where(entkw.IsActive(true))
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

func (s *Service) GetByID(id string) (*models.Keyword, error) {
	row, err := s.db.Client.Keyword.Get(context.Background(), id)
	if err != nil {
		return nil, err
	}
	m := mapKeyword(row, nil)
	return &m, nil
}

func (s *Service) Create(text string, category *string) (*models.Keyword, error) {
	now := time.Now().UTC()
	builder := s.db.Client.Keyword.Create().
		SetText(text).
		SetIsActive(true).
		SetCreatedAt(now).
		SetUpdatedAt(now)
	if category != nil {
		builder = builder.SetCategory(*category)
	}

	row, err := builder.Save(context.Background())
	if err != nil {
		return nil, err
	}
	m := mapKeyword(row, nil)
	return &m, nil
}

func (s *Service) Update(id string, text *string, category *string) (*models.Keyword, error) {
	now := time.Now().UTC()
	builder := s.db.Client.Keyword.UpdateOneID(id).SetUpdatedAt(now)
	if text != nil {
		builder = builder.SetText(*text)
	}
	if category != nil {
		builder = builder.SetCategory(*category)
	}

	if _, err := builder.Save(context.Background()); err != nil {
		return nil, err
	}
	return s.GetByID(id)
}

func (s *Service) Delete(id string) error {
	return s.db.Client.Keyword.DeleteOneID(id).Exec(context.Background())
}

func (s *Service) Toggle(id string) (*models.Keyword, error) {
	ctx := context.Background()
	row, err := s.db.Client.Keyword.Get(ctx, id)
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
