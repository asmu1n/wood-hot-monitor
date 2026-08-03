package hotspot

import "context"

type Repository interface {
	FindAll(ctx context.Context, params GetAllParams) ([]Hotspot, int, error)
	FindByID(ctx context.Context, id string) (*Hotspot, error)
	Search(ctx context.Context, params SearchParams) ([]Hotspot, int, error)
	Upsert(ctx context.Context, h Hotspot) (id string, isNew bool, err error)
	DeleteById(ctx context.Context, id string) error
	Delete(ctx context.Context, params DeleteParams) (int, error)
	CountByDeleteParams(ctx context.Context, params DeleteParams) (int, error)

	GetStatus(ctx context.Context) (*Status, error)
	GetNotifications(ctx context.Context, limit int) ([]Hotspot, error)
	UnreadCount(ctx context.Context) (int, error)
	MarkRead(ctx context.Context, id string) error
	MarkAllRead(ctx context.Context) error
}
