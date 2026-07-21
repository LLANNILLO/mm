package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/llannillo/mm/internal/shared/events"
	"github.com/llannillo/mm/internal/shared/outbox"
	store "github.com/llannillo/mm/modules/ticketing/internal/adapters/driven/postgres/generated"
	"github.com/llannillo/mm/modules/ticketing/internal/domain"
)

type EventRepository struct {
	queries *store.Queries
	uow     *UnitOfWork
}

func NewEventRepository(q *store.Queries, uow *UnitOfWork) *EventRepository {
	return &EventRepository{queries: q, uow: uow}
}

func (r *EventRepository) Insert(ctx context.Context, e *domain.Event) error {
	startsAtUtc := pgtype.Timestamptz{Time: e.StartsAtUtc(), Valid: true}
	endsAtUtc := pgtype.Timestamptz{}
	if e.EndsAtUtc() != nil {
		endsAtUtc = pgtype.Timestamptz{Time: *e.EndsAtUtc(), Valid: true}
	}

	err := r.queries.InsertEvent(ctx, store.InsertEventParams{
		ID:          e.ID(),
		Title:       e.Title(),
		Description: e.Description(),
		Location:    e.Location(),
		StartsAtUtc: startsAtUtc,
		EndsAtUtc:   endsAtUtc,
		Cancelled:   e.Cancelled(),
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
	return rehydrateTicketingEvent(row), nil
}

func rehydrateTicketingEvent(row store.TicketingEvent) *domain.Event {
	var endsAtUtc *time.Time
	if row.EndsAtUtc.Valid {
		t := row.EndsAtUtc.Time
		endsAtUtc = &t
	}
	return domain.RehydrateEvent(row.ID, row.Title, row.Description, row.Location, row.StartsAtUtc.Time, endsAtUtc, row.Cancelled)
}

func (r *EventRepository) Update(ctx context.Context, e *domain.Event) error {
	return r.uow.WithTx(ctx, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		for _, ev := range e.DomainEvents() {
			switch ev.(type) {
			case domain.EventRescheduledDomainEvent:
				startsAtUtc := pgtype.Timestamptz{Time: e.StartsAtUtc(), Valid: true}
				endsAtUtc := pgtype.Timestamptz{}
				if e.EndsAtUtc() != nil {
					endsAtUtc = pgtype.Timestamptz{Time: *e.EndsAtUtc(), Valid: true}
				}
				if err := q.UpdateEventSchedule(ctx, store.UpdateEventScheduleParams{
					StartsAtUtc: startsAtUtc,
					EndsAtUtc:   endsAtUtc,
					ID:          e.ID(),
				}); err != nil {
					return fmt.Errorf("update event schedule: %w", err)
				}
			case domain.EventCancelledDomainEvent:
				if err := q.CancelEvent(ctx, e.ID()); err != nil {
					return fmt.Errorf("cancel event: %w", err)
				}
			}
		}

		domainEvents := e.DomainEvents()
		e.ClearDomainEvents()
		return outbox.InsertMessages(ctx, tx, schema, domainEvents)
	})
}

// ArchiveTickets archives every non-archived ticket for the event and marks
// the event's tickets as archived — atomically. Called by the
// EventCancelled domain event handler.
func (r *EventRepository) ArchiveTickets(ctx context.Context, eventID uuid.UUID) error {
	return r.uow.WithTx(ctx, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		eventRow, err := q.GetEventByID(ctx, eventID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrEventNotFound
			}
			return fmt.Errorf("get event: %w", err)
		}
		event := rehydrateTicketingEvent(eventRow)

		ticketRows, err := q.GetTicketsByEventID(ctx, eventID)
		if err != nil {
			return fmt.Errorf("get tickets for event: %w", err)
		}

		var domainEvents []events.DomainEvent
		for _, tr := range ticketRows {
			ticket := domain.RehydrateTicket(tr.ID, tr.CustomerID, tr.OrderID, tr.EventID, tr.TicketTypeID, tr.Code, tr.CreatedAtUtc.Time, tr.Archived)
			ticket.Archive()

			if err := q.UpdateTicketArchived(ctx, ticket.ID()); err != nil {
				return fmt.Errorf("archive ticket %s: %w", ticket.ID(), err)
			}
			domainEvents = append(domainEvents, ticket.DomainEvents()...)
		}

		event.TicketsArchived()
		domainEvents = append(domainEvents, event.DomainEvents()...)

		return outbox.InsertMessages(ctx, tx, schema, domainEvents)
	})
}

// RefundPayments refunds the remaining balance of every not-yet-fully-
// refunded payment for the event and marks the event's payments as
// refunded — atomically. Called by the EventCancelled domain event handler.
// Persists refund state only; the actual gateway call happens later, out of
// band, via the PaymentRefunded/PaymentPartiallyRefunded domain event
// handlers dispatched off the outbox.
func (r *EventRepository) RefundPayments(ctx context.Context, eventID uuid.UUID) error {
	return r.uow.WithTx(ctx, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		eventRow, err := q.GetEventByID(ctx, eventID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrEventNotFound
			}
			return fmt.Errorf("get event: %w", err)
		}
		event := rehydrateTicketingEvent(eventRow)

		paymentRows, err := q.GetPaymentsByEventID(ctx, eventID)
		if err != nil {
			return fmt.Errorf("get payments for event: %w", err)
		}

		var domainEvents []events.DomainEvent
		for _, pr := range paymentRows {
			var amountRefunded *int64
			if pr.AmountRefunded.Valid {
				v := pr.AmountRefunded.Int64
				amountRefunded = &v
			}
			var refundedAtUtc *time.Time
			if pr.RefundedAtUtc.Valid {
				t := pr.RefundedAtUtc.Time
				refundedAtUtc = &t
			}
			payment := domain.RehydratePayment(pr.ID, pr.OrderID, pr.TransactionID, pr.Amount, pr.Currency, amountRefunded, pr.CreatedAtUtc.Time, refundedAtUtc)

			remaining := payment.Amount()
			if payment.AmountRefunded() != nil {
				remaining -= *payment.AmountRefunded()
			}
			if remaining <= 0 {
				continue
			}

			if err := payment.Refund(remaining); err != nil {
				return fmt.Errorf("refund payment %s: %w", payment.ID(), err)
			}

			amountRefundedParam := pgtype.Int8{}
			if payment.AmountRefunded() != nil {
				amountRefundedParam = pgtype.Int8{Int64: *payment.AmountRefunded(), Valid: true}
			}
			refundedAtUtcParam := pgtype.Timestamptz{}
			if payment.RefundedAtUtc() != nil {
				refundedAtUtcParam = pgtype.Timestamptz{Time: *payment.RefundedAtUtc(), Valid: true}
			}
			if err := q.UpdatePaymentRefund(ctx, store.UpdatePaymentRefundParams{
				AmountRefunded: amountRefundedParam,
				RefundedAtUtc:  refundedAtUtcParam,
				ID:             payment.ID(),
			}); err != nil {
				return fmt.Errorf("persist refund %s: %w", payment.ID(), err)
			}

			domainEvents = append(domainEvents, payment.DomainEvents()...)
		}

		event.PaymentsRefunded()
		domainEvents = append(domainEvents, event.DomainEvents()...)

		return outbox.InsertMessages(ctx, tx, schema, domainEvents)
	})
}
