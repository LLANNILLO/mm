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

type TicketRepository struct {
	queries *store.Queries
}

func NewTicketRepository(q *store.Queries) *TicketRepository {
	return &TicketRepository{queries: q}
}

func (r *TicketRepository) Insert(ctx context.Context, t *domain.Ticket) error {
	createdAtUtc := pgtype.Timestamptz{Time: t.CreatedAtUtc(), Valid: true}
	err := r.queries.InsertTicket(ctx, store.InsertTicketParams{
		ID:           t.ID(),
		CustomerID:   t.CustomerID(),
		OrderID:      t.OrderID(),
		EventID:      t.EventID(),
		TicketTypeID: t.TicketTypeID(),
		Code:         t.Code(),
		CreatedAtUtc: createdAtUtc,
		Archived:     t.Archived(),
	})
	if err != nil {
		return fmt.Errorf("insert ticket: %w", err)
	}
	t.ClearDomainEvents()
	return nil
}

func (r *TicketRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
	row, err := r.queries.GetTicketByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("ticket not found")
		}
		return nil, fmt.Errorf("get ticket: %w", err)
	}
	return domain.RehydrateTicket(row.ID, row.CustomerID, row.OrderID, row.EventID, row.TicketTypeID, row.Code, row.CreatedAtUtc.Time, row.Archived), nil
}
