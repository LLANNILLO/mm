package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	store "github.com/llannillo/mm/modules/ticketing/internal/adapters/driven/postgres/generated"
)

// EventStatisticsRepository writes directly through store.Queries — no
// UnitOfWork. Each call is a single statement invoked from an
// outbox.Idempotent-wrapped projection handler, which already guarantees
// at-least-once, deduplicated delivery; there is nothing else to coordinate
// a transaction with.
type EventStatisticsRepository struct {
	queries *store.Queries
}

func NewEventStatisticsRepository(q *store.Queries) *EventStatisticsRepository {
	return &EventStatisticsRepository{queries: q}
}

func (r *EventStatisticsRepository) EnsureRow(ctx context.Context, eventID uuid.UUID) error {
	if err := r.queries.EnsureEventStatisticsRow(ctx, eventID); err != nil {
		return fmt.Errorf("ensure event statistics row: %w", err)
	}
	return nil
}

func (r *EventStatisticsRepository) IncrementTicketsSold(ctx context.Context, eventID uuid.UUID) error {
	if err := r.queries.IncrementEventStatisticsTicketsSold(ctx, eventID); err != nil {
		return fmt.Errorf("increment tickets sold: %w", err)
	}
	return nil
}

func (r *EventStatisticsRepository) IncrementAttendeesCheckedIn(ctx context.Context, eventID uuid.UUID) error {
	if err := r.queries.IncrementEventStatisticsAttendeesCheckedIn(ctx, eventID); err != nil {
		return fmt.Errorf("increment attendees checked in: %w", err)
	}
	return nil
}

func (r *EventStatisticsRepository) AppendDuplicateCheckIn(ctx context.Context, eventID uuid.UUID, ticketCode string) error {
	if err := r.queries.AppendEventStatisticsDuplicateCheckIn(ctx, store.AppendEventStatisticsDuplicateCheckInParams{
		EventID:    eventID,
		TicketCode: ticketCode,
	}); err != nil {
		return fmt.Errorf("append duplicate check-in: %w", err)
	}
	return nil
}

func (r *EventStatisticsRepository) AppendInvalidCheckIn(ctx context.Context, eventID uuid.UUID, ticketCode string) error {
	if err := r.queries.AppendEventStatisticsInvalidCheckIn(ctx, store.AppendEventStatisticsInvalidCheckInParams{
		EventID:    eventID,
		TicketCode: ticketCode,
	}); err != nil {
		return fmt.Errorf("append invalid check-in: %w", err)
	}
	return nil
}
