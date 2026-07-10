package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	store "github.com/llannillo/mm/modules/ticketing/internal/adapters/driven/postgres/generated"
	"github.com/llannillo/mm/modules/ticketing/internal/domain"
)

type TicketTypeRepository struct {
	queries *store.Queries
}

func NewTicketTypeRepository(q *store.Queries) *TicketTypeRepository {
	return &TicketTypeRepository{queries: q}
}

func (r *TicketTypeRepository) Insert(ctx context.Context, tt *domain.TicketType) error {
	err := r.queries.InsertTicketType(ctx, store.InsertTicketTypeParams{
		ID:                tt.ID(),
		EventID:           tt.EventID(),
		Name:              tt.Name(),
		Price:             tt.Price(),
		Currency:          tt.Currency(),
		Quantity:          tt.Quantity(),
		AvailableQuantity: tt.AvailableQuantity(),
	})
	if err != nil {
		return fmt.Errorf("insert ticket type: %w", err)
	}
	return nil
}

func (r *TicketTypeRepository) InsertBatch(ctx context.Context, ticketTypes []*domain.TicketType) error {
	for _, tt := range ticketTypes {
		if err := r.Insert(ctx, tt); err != nil {
			return err
		}
	}
	return nil
}

func (r *TicketTypeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.TicketType, error) {
	row, err := r.queries.GetTicketTypeByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTicketTypeNotFound
		}
		return nil, fmt.Errorf("get ticket type: %w", err)
	}
	return domain.RehydrateTicketType(row.ID, row.EventID, row.Name, row.Price, row.Currency, row.Quantity, row.AvailableQuantity), nil
}

func (r *TicketTypeRepository) Update(ctx context.Context, tt *domain.TicketType) error {
	err := r.queries.UpdateTicketTypePrice(ctx, store.UpdateTicketTypePriceParams{
		Price: tt.Price(),
		ID:    tt.ID(),
	})
	if err != nil {
		return fmt.Errorf("update ticket type price: %w", err)
	}
	tt.ClearDomainEvents()
	return nil
}
