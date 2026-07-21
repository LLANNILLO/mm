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

	// ArchiveTickets archives every non-archived ticket for the event and
	// marks the event's tickets as archived — atomically. Safe to retry:
	// already-archived tickets are excluded from the batch.
	ArchiveTickets(ctx context.Context, eventID uuid.UUID) error

	// RefundPayments refunds the remaining balance of every not-yet-fully-
	// refunded payment for the event and marks the event's payments as
	// refunded — atomically. Safe to retry: fully-refunded payments are
	// excluded from the batch.
	RefundPayments(ctx context.Context, eventID uuid.UUID) error
}
