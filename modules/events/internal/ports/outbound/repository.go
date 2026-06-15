package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/llannillo/mm/modules/events/internal/domain"
)

type EventRepository interface {
	Insert(ctx context.Context, event *domain.Event) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error)
	Update(ctx context.Context, event *domain.Event) error
}

type CategoryRepository interface {
	Insert(ctx context.Context, c *domain.Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error)
	Update(ctx context.Context, c *domain.Category) error
}

type TicketTypeRepository interface {
	Insert(ctx context.Context, tt *domain.TicketType) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.TicketType, error)
	Update(ctx context.Context, tt *domain.TicketType) error
	ExistsByEventID(ctx context.Context, eventID uuid.UUID) (bool, error)
}
