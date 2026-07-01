package hotspot

import "wood-hot-monitor/internal/models"

type Service struct{}

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

func (s *Service) GetAll(params GetAllParams) (*models.PaginatedResult[models.Hotspot], error) {
	return &models.PaginatedResult[models.Hotspot]{
		Data:  []models.Hotspot{},
		Total: 0,
		Page:  params.Page,
		Limit: params.Limit,
	}, nil
}

func (s *Service) GetByID(id string) (*models.Hotspot, error) {
	return nil, nil
}

func (s *Service) GetStatus() (*models.Status, error) {
	return &models.Status{
		Total:    0,
		Today:    0,
		Urgent:   0,
		BySource: map[string]int{},
	}, nil
}

func (s *Service) Delete(id string) error {
	return nil
}

func (s *Service) Search(query string, sources []string) ([]models.Hotspot, error) {
	return []models.Hotspot{}, nil
}

func (s *Service) Check() error {
	return nil
}
