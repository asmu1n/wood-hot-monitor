package setting

type Service struct{}

func (s *Service) GetAll() (map[string]string, error) {
	return map[string]string{}, nil
}

func (s *Service) GetByKey(key string) (string, error) {
	return "", nil
}

func (s *Service) Update(key string, value string) error {
	return nil
}

func (s *Service) BulkUpdate(settings map[string]string) error {
	return nil
}
