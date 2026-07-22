package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	store "github.com/llannillo/mm/modules/ticketing/internal/adapters/driven/postgres/generated"
	geteventstatistics "github.com/llannillo/mm/modules/ticketing/internal/app/queries/get_event_statistics"
	"github.com/llannillo/mm/modules/ticketing/internal/domain"
)

type EventStatisticsReader struct {
	queries *store.Queries
}

func NewEventStatisticsReader(q *store.Queries) *EventStatisticsReader {
	return &EventStatisticsReader{queries: q}
}

func (r *EventStatisticsReader) GetEventStatistics(ctx context.Context, eventID uuid.UUID) (*geteventstatistics.Response, error) {
	row, err := r.queries.GetEventStatisticsByEventID(ctx, eventID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEventStatisticsNotFound
		}
		return nil, fmt.Errorf("get event statistics: %w", err)
	}
	return &geteventstatistics.Response{
		EventID:                 row.EventID,
		TicketsSold:             row.TicketsSold,
		AttendeesCheckedIn:      row.AttendeesCheckedIn,
		DuplicateCheckInTickets: row.DuplicateCheckInTickets,
		InvalidCheckInTickets:   row.InvalidCheckInTickets,
	}, nil
}
