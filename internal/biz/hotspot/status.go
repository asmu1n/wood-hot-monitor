package hotspot

import (
	"context"
	"time"
	"wood-hot-monitor/ent"
	hsmodel "wood-hot-monitor/ent/hotspot"
	"wood-hot-monitor/internal/core/models"
)

func (s *Service) GetStatus() (*models.Status, error) {
	ctx := context.Background()

	total, _ := s.db.Client.Hotspot.Query().Count(ctx)

	todayStart := time.Now().UTC().Truncate(24 * time.Hour)
	today, _ := s.db.Client.Hotspot.Query().
		Where(hsmodel.CreatedAtGTE(todayStart)).Count(ctx)

	urgent, _ := s.db.Client.Hotspot.Query().
		Where(hsmodel.ImportanceEQ("urgent")).Count(ctx)

	bySource := make(map[string]int)
	var sourceCounts []struct {
		Source string `json:"source"`
		Count  int    `json:"count"`
	}
	err := s.db.Client.Hotspot.Query().
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

func (s *Service) GetNotifications(limit int) ([]models.Hotspot, error) {
	ctx := context.Background()
	if limit <= 0 {
		limit = 10
	}

	rows, err := s.db.Client.Hotspot.Query().
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

func (s *Service) UnreadCount() (int, error) {
	return s.db.Client.Hotspot.Query().
		Where(hsmodel.IsNotified(true), hsmodel.IsRead(false)).
		Count(context.Background())
}

func (s *Service) MarkRead(id string) error {
	return s.db.Client.Hotspot.UpdateOneID(id).
		SetIsRead(true).
		Exec(context.Background())
}

func (s *Service) MarkAllRead() error {
	return s.db.Client.Hotspot.Update().
		Where(hsmodel.IsNotified(true), hsmodel.IsRead(false)).
		SetIsRead(true).
		Exec(context.Background())
}
