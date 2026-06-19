package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/llannillo/mm/modules/ticketing/internal/domain"
)

type TicketTypeRepository interface {
	Insert(ctx context.Context, ticketType *domain.TicketType) error
	InsertBatch(ctx context.Context, ticketTypes []*domain.TicketType) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.TicketType, error)
	Update(ctx context.Context, ticketType *domain.TicketType) error
}
