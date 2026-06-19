package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	store "github.com/llannillo/mm/modules/ticketing/internal/adapters/driven/postgres/generated"
	"github.com/llannillo/mm/modules/ticketing/internal/domain"
)

type EventRepository struct {
	queries *store.Queries
}

func NewEventRepository(q *store.Queries) *EventRepository {
	return &EventRepository{queries: q}
}

func (r *EventRepository) Insert(ctx context.Context, e *domain.Event) error {
	startsAtUtc := pgtype.Timestamptz{Time: e.StartsAtUtc, Valid: true}
	endsAtUtc := pgtype.Timestamptz{}
	if e.EndsAtUtc != nil {
		endsAtUtc = pgtype.Timestamptz{Time: *e.EndsAtUtc, Valid: true}
	}

	err := r.queries.InsertEvent(ctx, store.InsertEventParams{
		ID:          e.ID,
		Title:       e.Title,
		Description: e.Description,
		Location:    e.Location,
		StartsAtUtc: startsAtUtc,
		EndsAtUtc:   endsAtUtc,
		Cancelled:   e.Cancelled,
	})
	if err != nil {
		return fmt.Errorf("insert event: %w", err)
	}
	return nil
}

func (r *EventRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	row, err := r.queries.GetEventByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEventNotFound
		}
		return nil, fmt.Errorf("get event: %w", err)
	}

	e := &domain.Event{
		ID:          row.ID,
		Title:       row.Title,
		Description: row.Description,
		Location:    row.Location,
		StartsAtUtc: row.StartsAtUtc.Time,
		Cancelled:   row.Cancelled,
	}
	if row.EndsAtUtc.Valid {
		t := row.EndsAtUtc.Time
		e.EndsAtUtc = &t
	}
	return e, nil
}

func (r *EventRepository) Update(ctx context.Context, e *domain.Event) error {
	for _, ev := range e.DomainEvents() {
		switch ev.(type) {
		case domain.EventRescheduledDomainEvent:
			startsAtUtc := pgtype.Timestamptz{Time: e.StartsAtUtc, Valid: true}
			endsAtUtc := pgtype.Timestamptz{}
			if e.EndsAtUtc != nil {
				endsAtUtc = pgtype.Timestamptz{Time: *e.EndsAtUtc, Valid: true}
			}
			if err := r.queries.UpdateEventSchedule(ctx, store.UpdateEventScheduleParams{
				StartsAtUtc: startsAtUtc,
				EndsAtUtc:   endsAtUtc,
				ID:          e.ID,
			}); err != nil {
				return fmt.Errorf("update event schedule: %w", err)
			}
		case domain.EventCancelledDomainEvent:
			if err := r.queries.CancelEvent(ctx, e.ID); err != nil {
				return fmt.Errorf("cancel event: %w", err)
			}
		}
	}
	e.ClearDomainEvents()
	return nil
}
