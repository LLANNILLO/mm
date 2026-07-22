package outbound

import (
	"context"

	"github.com/google/uuid"
)

// EventStatisticsRepository maintains the ticketing.event_statistics
// materialized view — a denormalized read model kept in sync by projection
// handlers reacting to ticket domain events, not by the write-side aggregates.
type EventStatisticsRepository interface {
	EnsureRow(ctx context.Context, eventID uuid.UUID) error
	IncrementTicketsSold(ctx context.Context, eventID uuid.UUID) error
	IncrementAttendeesCheckedIn(ctx context.Context, eventID uuid.UUID) error
	AppendDuplicateCheckIn(ctx context.Context, eventID uuid.UUID, ticketCode string) error
	AppendInvalidCheckIn(ctx context.Context, eventID uuid.UUID, ticketCode string) error
}
