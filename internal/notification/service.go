package notification

import "wood-hot-monitor/internal/models"

type Service struct{}

type GetAllParams struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	UnreadOnly bool `json:"unreadOnly"`
}

func (s *Service) GetAll(params GetAllParams) (*models.PaginatedResult[models.Notification], error) {
	return &models.PaginatedResult[models.Notification]{
		Data:  []models.Notification{},
		Total: 0,
		Page:  params.Page,
		Limit: params.Limit,
	}, nil
}

func (s *Service) MarkAsRead(id string) error {
	return nil
}

func (s *Service) MarkAllAsRead() error {
	return nil
}

func (s *Service) Delete(id string) error {
	return nil
}

func (s *Service) ClearAll() error {
	return nil
}
