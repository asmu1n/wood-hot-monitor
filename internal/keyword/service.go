package keyword

import "wood-hot-monitor/internal/models"

type Service struct{}

func (s *Service) GetAll() ([]models.Keyword, error) {
	return []models.Keyword{}, nil
}

func (s *Service) GetByID(id string) (*models.Keyword, error) {
	return nil, nil
}

func (s *Service) Create(text string, category *string) (*models.Keyword, error) {
	return nil, nil
}

func (s *Service) Update(id string, text *string, category *string) (*models.Keyword, error) {
	return nil, nil
}

func (s *Service) Delete(id string) error {
	return nil
}

func (s *Service) Toggle(id string) (*models.Keyword, error) {
	return nil, nil
}
