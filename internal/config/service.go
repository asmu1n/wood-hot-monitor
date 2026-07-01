package config

import "wood-hot-monitor/internal/models"

type Service struct{}

func (s *Service) Get() (*models.AppConfig, error) {
	return &models.AppConfig{
		CustomOptions: map[string]string{},
	}, nil
}

func (s *Service) Update(cfg models.AppConfig) error {
	return nil
}
