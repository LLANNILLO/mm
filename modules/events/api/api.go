package eventsapi

import (
	"context"

	"github.com/google/uuid"
)

type EventsAPI interface {
	GetTicketType(ctx context.Context, id uuid.UUID) (*TicketTypeResponse, error)
}

type TicketTypeResponse struct {
	ID       uuid.UUID
	EventID  uuid.UUID
	Name     string
	Price    int64
	Currency string
	Quantity int64
}
