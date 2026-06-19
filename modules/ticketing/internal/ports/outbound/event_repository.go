package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/llannillo/mm/modules/ticketing/internal/domain"
)

type EventRepository interface {
	Insert(ctx context.Context, event *domain.Event) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error)
	Update(ctx context.Context, event *domain.Event) error
}
