package notification

import (
	"context"
	"time"

	"wood-hot-monitor/ent"
	entnotif "wood-hot-monitor/ent/notification"
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
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	UnreadOnly bool `json:"unreadOnly"`
}

func (s *Service) GetAll(params GetAllParams) (*models.PaginatedResult[models.Notification], error) {
	ctx := context.Background()

	query := s.db.Client.Notification.Query()
	if params.UnreadOnly {
		query = query.Where(entnotif.IsRead(false))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}

	offset := (params.Page - 1) * params.Limit
	rows, err := query.
		Order(ent.Desc(entnotif.FieldCreatedAt)).
		Limit(params.Limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, err
	}

	data := make([]models.Notification, len(rows))
	for i, row := range rows {
		data[i] = mapNotification(row)
	}

	return &models.PaginatedResult[models.Notification]{
		Data:  data,
		Total: total,
		Page:  params.Page,
		Limit: params.Limit,
	}, nil
}

func (s *Service) MarkAsRead(id string) error {
	return s.db.Client.Notification.UpdateOneID(id).
		SetIsRead(true).
		Exec(context.Background())
}

func (s *Service) MarkAllAsRead() error {
	_, err := s.db.Client.Notification.Update().
		Where(entnotif.IsRead(false)).
		SetIsRead(true).
		Save(context.Background())
	return err
}

func (s *Service) Delete(id string) error {
	return s.db.Client.Notification.DeleteOneID(id).Exec(context.Background())
}

func (s *Service) ClearAll() error {
	_, err := s.db.Client.Notification.Delete().Exec(context.Background())
	return err
}

func (s *Service) Create(typ, title, content string, hotspotID *string) (*models.Notification, error) {
	builder := s.db.Client.Notification.Create().
		SetType(typ).
		SetTitle(title).
		SetContent(content).
		SetIsRead(false).
		SetCreatedAt(time.Now().UTC())
	if hotspotID != nil {
		builder = builder.SetHotspotID(*hotspotID)
	}

	row, err := builder.Save(context.Background())
	if err != nil {
		return nil, err
	}

	m := mapNotification(row)
	return &m, nil
}

func mapNotification(row *ent.Notification) models.Notification {
	return models.Notification{
		ID:        row.ID,
		Type:      row.Type,
		Title:     row.Title,
		Content:   row.Content,
		IsRead:    row.IsRead,
		HotSpotID: row.HotspotID,
		CreatedAt: row.CreatedAt.Format(time.DateTime),
	}
}
