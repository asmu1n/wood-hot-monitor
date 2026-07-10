package hotspot

import "context"

// Repository 热点持久化接口
type Repository interface {
	FindAll(ctx context.Context, filter Filter) ([]Hotspot, int, error)
	FindByID(ctx context.Context, id string) (*Hotspot, error)
	Search(ctx context.Context, filter SearchFilter) ([]Hotspot, int, error)
	Upsert(ctx context.Context, h Hotspot) (id string, isNew bool, err error)
	Delete(ctx context.Context, id string) error

	GetStatus(ctx context.Context) (*Status, error)
	GetNotifications(ctx context.Context, limit int) ([]Hotspot, error)
	UnreadCount(ctx context.Context) (int, error)
	MarkRead(ctx context.Context, id string) error
	MarkAllRead(ctx context.Context) error
}
