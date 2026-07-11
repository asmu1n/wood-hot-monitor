package keyword

import "context"

type Repository interface {
	FindAll(ctx context.Context, activeOnly bool) ([]Keyword, error)
	FindByID(ctx context.Context, id string) (*Keyword, error)
	Create(ctx context.Context, text string, category *string) (*Keyword, error)
	Update(ctx context.Context, id string, text *string, category *string) (*Keyword, error)
	Delete(ctx context.Context, id string) error
	Toggle(ctx context.Context, id string) (*Keyword, error)
}
