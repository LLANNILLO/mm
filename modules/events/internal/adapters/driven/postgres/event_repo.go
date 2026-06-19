package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/llannillo/mm/internal/shared/events"
	"github.com/llannillo/mm/modules/events/internal/domain"
	store "github.com/llannillo/mm/modules/events/internal/adapters/driven/postgres/generated"
)

type EventRepository struct {
	queries    *store.Queries
	dispatcher *events.Dispatcher
}

func NewEventRepository(q *store.Queries, d *events.Dispatcher) *EventRepository {
	return &EventRepository{queries: q, dispatcher: d}
}

func (r *EventRepository) Insert(ctx context.Context, event *domain.Event) error {
	_, err := r.queries.InsertEvent(ctx, store.InsertEventParams{
		ID:          event.ID,
		CategoryID:  event.CategoryID,
		Title:       event.Title,
		Description: event.Description,
		Location:    event.Location,
		StartsAtUtc: event.StartsAtUtc,
		EndsAtUtc:   event.EndsAtUtc,
		Status:      event.Status,
	})
	if err != nil {
		return fmt.Errorf("insert event: %w", err)
	}
	domainEvents := event.DomainEvents()
	event.ClearDomainEvents()
	return r.dispatcher.Dispatch(ctx, domainEvents)
}

func (r *EventRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	row, err := r.queries.SelectEventForUpdate(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEventNotFound
		}
		return nil, fmt.Errorf("get event by id: %w", err)
	}
	return rehydrateEvent(row), nil
}

func (r *EventRepository) Update(ctx context.Context, event *domain.Event) error {
	for _, e := range event.DomainEvents() {
		var err error
		switch e.(type) {
		case domain.EventCancelledDomainEvent:
			err = r.queries.UCancelEvent(ctx, event.ID)
		case domain.EventPublishedDomainEvent:
			err = r.queries.UPublishEvent(ctx, event.ID)
		case domain.EventRescheduledDomainEvent:
			err = r.queries.URescheduleEvent(ctx, store.URescheduleEventParams{
				ID:         event.ID,
				StartsDate: event.StartsAtUtc,
				EndsDate:   event.EndsAtUtc,
			})
		}
		if err != nil {
			return err
		}
	}
	domainEvents := event.DomainEvents()
	event.ClearDomainEvents()
	return r.dispatcher.Dispatch(ctx, domainEvents)
}

func rehydrateEvent(row store.SelectEventForUpdateRow) *domain.Event {
	return &domain.Event{
		ID:          row.ID,
		CategoryID:  row.CategoryID,
		Title:       row.Title,
		Description: row.Description,
		Location:    row.Location,
		StartsAtUtc: row.StartsAtUtc,
		EndsAtUtc:   row.EndsAtUtc,
		Status:      row.Status,
	}
}
