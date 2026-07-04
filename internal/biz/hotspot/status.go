package hotspot

import (
	"context"
	"time"
	"wood-hot-monitor/ent"
	hsmodel "wood-hot-monitor/ent/hotspot"
	"wood-hot-monitor/internal/core/models"
)

func (s *Service) GetStatus(ctx context.Context) (*models.Status, error) {
	total, _ := s.client.Hotspot.Query().Count(ctx)

	todayStart := time.Now().UTC().Truncate(24 * time.Hour)
	today, _ := s.client.Hotspot.Query().
		Where(hsmodel.CreatedAtGTE(todayStart)).Count(ctx)

	urgent, _ := s.client.Hotspot.Query().
		Where(hsmodel.ImportanceEQ("urgent")).Count(ctx)

	bySource := make(map[string]int)
	var sourceCounts []struct {
		Source string `json:"source"`
		Count  int    `json:"count"`
	}
	err := s.client.Hotspot.Query().
		GroupBy(hsmodel.FieldSource).
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

func (s *Service) GetNotifications(ctx context.Context, limit int) ([]models.Hotspot, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := s.client.Hotspot.Query().
		Where(hsmodel.IsNotified(true)).
		Order(ent.Desc(hsmodel.FieldNotifiedAt)).
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

func (s *Service) UnreadCount(ctx context.Context) (int, error) {
	return s.client.Hotspot.Query().
		Where(hsmodel.IsNotified(true), hsmodel.IsRead(false)).
		Count(ctx)
}

func (s *Service) MarkRead(ctx context.Context, id string) error {
	return s.client.Hotspot.UpdateOneID(id).
		SetIsRead(true).
		Exec(ctx)
}

func (s *Service) MarkAllRead(ctx context.Context) error {
	return s.client.Hotspot.Update().
		Where(hsmodel.IsNotified(true), hsmodel.IsRead(false)).
		SetIsRead(true).
		Exec(ctx)
}
