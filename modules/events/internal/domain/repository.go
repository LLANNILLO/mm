package domain

import (
	"context"

	"github.com/google/uuid"
)

type EventRepository interface {
	Insert(ctx context.Context, event *Event) error
	GetByID(ctx context.Context, id uuid.UUID) (*Event, error)
	Update(ctx context.Context, event *Event) error
}

type CategoryRepository interface {
	Insert(ctx context.Context, c *Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*Category, error)
	Update(ctx context.Context, c *Category) error
}

type TicketTypeRepository interface {
	Insert(ctx context.Context, tt *TicketType) error
	GetByID(ctx context.Context, id uuid.UUID) (*TicketType, error)
	Update(ctx context.Context, tt *TicketType) error
	ExistsByEventID(ctx context.Context, eventID uuid.UUID) (bool, error)
}
