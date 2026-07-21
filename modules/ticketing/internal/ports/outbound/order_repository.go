package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/llannillo/mm/modules/ticketing/internal/domain"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)

	// IssueTickets creates one Ticket per unit of quantity across the order's
	// items and marks the order as having its tickets issued — atomically, in
	// a single transaction. Safe to retry: a already-issued order is a no-op.
	IssueTickets(ctx context.Context, orderID uuid.UUID) error
}
