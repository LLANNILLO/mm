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

type TicketTypeRepository struct {
	queries    *store.Queries
	dispatcher *events.Dispatcher
}

func NewTicketTypeRepository(q *store.Queries, d *events.Dispatcher) *TicketTypeRepository {
	return &TicketTypeRepository{queries: q, dispatcher: d}
}

func (r *TicketTypeRepository) Insert(ctx context.Context, tt *domain.TicketType) error {
	_, err := r.queries.InsertTicketType(ctx, store.InsertTicketTypeParams{
		ID:       tt.ID,
		EventID:  tt.EventID,
		Name:     tt.Name,
		Price:    tt.Price,
		Currency: tt.Currency,
		Quantity: tt.Quantity,
	})
	if err != nil {
		return fmt.Errorf("insert ticket type: %w", err)
	}
	domainEvents := tt.DomainEvents()
	tt.ClearDomainEvents()
	return r.dispatcher.Dispatch(ctx, domainEvents)
}

func (r *TicketTypeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.TicketType, error) {
	row, err := r.queries.SelectTicketTypeByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTicketTypeNotFound
		}
		return nil, fmt.Errorf("get ticket type by id: %w", err)
	}
	return &domain.TicketType{
		ID:       row.ID,
		EventID:  row.EventID,
		Name:     row.Name,
		Price:    row.Price,
		Currency: row.Currency,
		Quantity: row.Quantity,
	}, nil
}

func (r *TicketTypeRepository) Update(ctx context.Context, tt *domain.TicketType) error {
	for _, e := range tt.DomainEvents() {
		var err error
		switch e.(type) {
		case domain.TicketTypePriceChangedDomainEvent:
			err = r.queries.UpdateTicketTypePrice(ctx, store.UpdateTicketTypePriceParams{
				ID:    tt.ID,
				Price: tt.Price,
			})
		}
		if err != nil {
			return err
		}
	}
	domainEvents := tt.DomainEvents()
	tt.ClearDomainEvents()
	return r.dispatcher.Dispatch(ctx, domainEvents)
}

func (r *TicketTypeRepository) ExistsByEventID(ctx context.Context, eventID uuid.UUID) (bool, error) {
	exists, err := r.queries.SExistsTicketTypeByEventID(ctx, eventID)
	if err != nil {
		return false, fmt.Errorf("exists ticket type by event id: %w", err)
	}
	return exists, nil
}
