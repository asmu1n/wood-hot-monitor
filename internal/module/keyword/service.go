package keyword

import "context"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetAll(ctx context.Context, activeOnly bool) ([]Keyword, error) {
	return s.repo.FindAll(ctx, activeOnly)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Keyword, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, text string, category *string) (*Keyword, error) {
	return s.repo.Create(ctx, text, category)
}

func (s *Service) Update(ctx context.Context, id string, text *string, category *string) (*Keyword, error) {
	return s.repo.Update(ctx, id, text, category)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) Toggle(ctx context.Context, id string) (*Keyword, error) {
	return s.repo.Toggle(ctx, id)
}
