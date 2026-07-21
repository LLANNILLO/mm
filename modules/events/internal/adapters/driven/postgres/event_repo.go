package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/llannillo/mm/internal/shared/outbox"
	store "github.com/llannillo/mm/modules/events/internal/adapters/driven/postgres/generated"
	"github.com/llannillo/mm/modules/events/internal/domain"
)

type EventRepository struct {
	queries *store.Queries
	uow     *UnitOfWork
}

func NewEventRepository(q *store.Queries, uow *UnitOfWork) *EventRepository {
	return &EventRepository{queries: q, uow: uow}
}

func (r *EventRepository) Insert(ctx context.Context, event *domain.Event) error {
	return r.uow.WithTx(ctx, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		_, err := q.InsertEvent(ctx, store.InsertEventParams{
			ID:          event.ID(),
			CategoryID:  event.CategoryID(),
			Title:       event.Title(),
			Description: event.Description(),
			Location:    event.Location(),
			StartsAtUtc: event.StartsAtUtc(),
			EndsAtUtc:   event.EndsAtUtc(),
			Status:      event.Status(),
		})
		if err != nil {
			return fmt.Errorf("insert event: %w", err)
		}

		domainEvents := event.DomainEvents()
		event.ClearDomainEvents()
		return outbox.InsertMessages(ctx, tx, schema, domainEvents)
	})
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
	return r.uow.WithTx(ctx, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		for _, e := range event.DomainEvents() {
			var err error
			switch e.(type) {
			case domain.EventCancelledDomainEvent:
				err = q.UCancelEvent(ctx, event.ID())
			case domain.EventPublishedDomainEvent:
				err = q.UPublishEvent(ctx, event.ID())
			case domain.EventRescheduledDomainEvent:
				err = q.URescheduleEvent(ctx, store.URescheduleEventParams{
					ID:         event.ID(),
					StartsDate: event.StartsAtUtc(),
					EndsDate:   event.EndsAtUtc(),
				})
			}
			if err != nil {
				return err
			}
		}

		domainEvents := event.DomainEvents()
		event.ClearDomainEvents()
		return outbox.InsertMessages(ctx, tx, schema, domainEvents)
	})
}

func rehydrateEvent(row store.SelectEventForUpdateRow) *domain.Event {
	return domain.RehydrateEvent(
		row.ID,
		row.CategoryID,
		row.Title,
		row.Description,
		row.Location,
		row.StartsAtUtc,
		row.EndsAtUtc,
		row.Status,
	)
}
