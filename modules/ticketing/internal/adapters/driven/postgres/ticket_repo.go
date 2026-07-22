package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/llannillo/mm/internal/shared/outbox"
	store "github.com/llannillo/mm/modules/ticketing/internal/adapters/driven/postgres/generated"
	"github.com/llannillo/mm/modules/ticketing/internal/domain"
)

type TicketRepository struct {
	queries *store.Queries
	uow     *UnitOfWork
}

func NewTicketRepository(q *store.Queries, uow *UnitOfWork) *TicketRepository {
	return &TicketRepository{queries: q, uow: uow}
}

func (r *TicketRepository) Insert(ctx context.Context, t *domain.Ticket) error {
	return r.uow.WithTx(ctx, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		createdAtUtc := pgtype.Timestamptz{Time: t.CreatedAtUtc(), Valid: true}
		usedAtUtc := pgtype.Timestamptz{}
		if t.UsedAtUtc() != nil {
			usedAtUtc = pgtype.Timestamptz{Time: *t.UsedAtUtc(), Valid: true}
		}
		if err := q.InsertTicket(ctx, store.InsertTicketParams{
			ID:           t.ID(),
			CustomerID:   t.CustomerID(),
			OrderID:      t.OrderID(),
			EventID:      t.EventID(),
			TicketTypeID: t.TicketTypeID(),
			Code:         t.Code(),
			CreatedAtUtc: createdAtUtc,
			Archived:     t.Archived(),
			UsedAtUtc:    usedAtUtc,
		}); err != nil {
			return fmt.Errorf("insert ticket: %w", err)
		}

		domainEvents := t.DomainEvents()
		t.ClearDomainEvents()
		return outbox.InsertMessages(ctx, tx, schema, domainEvents)
	})
}

func (r *TicketRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
	row, err := r.queries.GetTicketByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("ticket not found")
		}
		return nil, fmt.Errorf("get ticket: %w", err)
	}
	return rehydrateTicketingTicket(row), nil
}

func rehydrateTicketingTicket(row store.TicketingTicket) *domain.Ticket {
	var usedAtUtc *time.Time
	if row.UsedAtUtc.Valid {
		t := row.UsedAtUtc.Time
		usedAtUtc = &t
	}
	return domain.RehydrateTicket(row.ID, row.CustomerID, row.OrderID, row.EventID, row.TicketTypeID, row.Code, row.CreatedAtUtc.Time, row.Archived, usedAtUtc)
}

// Update persists domain events raised on t. Only TicketCheckedInDomainEvent
// maps to a state change (used_at_utc); TicketCheckInDuplicateDomainEvent and
// TicketCheckInInvalidDomainEvent carry no state change of their own — they
// still flow to the outbox so the event_statistics projections can react.
func (r *TicketRepository) Update(ctx context.Context, t *domain.Ticket) error {
	return r.uow.WithTx(ctx, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		for _, ev := range t.DomainEvents() {
			switch ev.(type) {
			case domain.TicketCheckedInDomainEvent:
				usedAtUtc := pgtype.Timestamptz{Time: *t.UsedAtUtc(), Valid: true}
				if err := q.UpdateTicketCheckedIn(ctx, store.UpdateTicketCheckedInParams{
					UsedAtUtc: usedAtUtc,
					ID:        t.ID(),
				}); err != nil {
					return fmt.Errorf("update ticket checked in: %w", err)
				}
			}
		}

		domainEvents := t.DomainEvents()
		t.ClearDomainEvents()
		return outbox.InsertMessages(ctx, tx, schema, domainEvents)
	})
}
