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

type TicketTypeRepository struct {
	queries *store.Queries
	uow     *UnitOfWork
}

func NewTicketTypeRepository(q *store.Queries, uow *UnitOfWork) *TicketTypeRepository {
	return &TicketTypeRepository{queries: q, uow: uow}
}

func (r *TicketTypeRepository) Insert(ctx context.Context, tt *domain.TicketType) error {
	return r.uow.WithTx(ctx, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		_, err := q.InsertTicketType(ctx, store.InsertTicketTypeParams{
			ID:       tt.ID(),
			EventID:  tt.EventID(),
			Name:     tt.Name(),
			Price:    tt.Price(),
			Currency: tt.Currency(),
			Quantity: tt.Quantity(),
		})
		if err != nil {
			return fmt.Errorf("insert ticket type: %w", err)
		}

		domainEvents := tt.DomainEvents()
		tt.ClearDomainEvents()
		return outbox.InsertMessages(ctx, tx, schema, domainEvents)
	})
}

func (r *TicketTypeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.TicketType, error) {
	row, err := r.queries.SelectTicketTypeByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTicketTypeNotFound
		}
		return nil, fmt.Errorf("get ticket type by id: %w", err)
	}
	return domain.RehydrateTicketType(row.ID, row.EventID, row.Name, row.Price, row.Currency, row.Quantity), nil
}

func (r *TicketTypeRepository) Update(ctx context.Context, tt *domain.TicketType) error {
	return r.uow.WithTx(ctx, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		for _, e := range tt.DomainEvents() {
			var err error
			switch e.(type) {
			case domain.TicketTypePriceChangedDomainEvent:
				err = q.UpdateTicketTypePrice(ctx, store.UpdateTicketTypePriceParams{
					ID:    tt.ID(),
					Price: tt.Price(),
				})
			}
			if err != nil {
				return err
			}
		}

		domainEvents := tt.DomainEvents()
		tt.ClearDomainEvents()
		return outbox.InsertMessages(ctx, tx, schema, domainEvents)
	})
}

func (r *TicketTypeRepository) ExistsByEventID(ctx context.Context, eventID uuid.UUID) (bool, error) {
	exists, err := r.queries.SExistsTicketTypeByEventID(ctx, eventID)
	if err != nil {
		return false, fmt.Errorf("exists ticket type by event id: %w", err)
	}
	return exists, nil
}
