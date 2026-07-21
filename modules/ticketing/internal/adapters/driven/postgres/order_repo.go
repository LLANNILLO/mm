package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/llannillo/mm/internal/shared/outbox"
	store "github.com/llannillo/mm/modules/ticketing/internal/adapters/driven/postgres/generated"
	"github.com/llannillo/mm/modules/ticketing/internal/domain"
)

type OrderRepository struct {
	queries *store.Queries
	uow     *UnitOfWork
}

func NewOrderRepository(q *store.Queries, uow *UnitOfWork) *OrderRepository {
	return &OrderRepository{queries: q, uow: uow}
}

func (r *OrderRepository) Create(ctx context.Context, order *domain.Order) error {
	return r.uow.WithTx(ctx, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		// For each order item: lock the ticket_type row, validate and decrement available_quantity
		for _, item := range order.Items() {
			var availableQty int64
			row := tx.QueryRow(ctx,
				`SELECT available_quantity FROM ticketing.ticket_types WHERE id = $1 FOR UPDATE`,
				item.TicketTypeID(),
			)
			if err := row.Scan(&availableQty); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return domain.ErrTicketTypeNotFound
				}
				return fmt.Errorf("lock ticket type: %w", err)
			}

			if item.Quantity() > availableQty {
				return domain.ErrTicketTypeInsufficientQuantity
			}

			if err := q.DecrementTicketTypeQuantity(ctx, store.DecrementTicketTypeQuantityParams{
				AvailableQuantity: item.Quantity(),
				ID:                item.TicketTypeID(),
			}); err != nil {
				return fmt.Errorf("decrement ticket type quantity: %w", err)
			}
		}

		// Insert the order
		createdAtUtc := pgtype.Timestamptz{Time: order.CreatedAtUtc(), Valid: true}
		if err := q.InsertOrder(ctx, store.InsertOrderParams{
			ID:            order.ID(),
			CustomerID:    order.CustomerID(),
			Status:        string(order.Status()),
			TotalPrice:    order.TotalPrice(),
			Currency:      order.Currency(),
			TicketsIssued: order.TicketsIssued(),
			CreatedAtUtc:  createdAtUtc,
		}); err != nil {
			return fmt.Errorf("insert order: %w", err)
		}

		// Insert all order items
		for _, item := range order.Items() {
			if err := q.InsertOrderItem(ctx, store.InsertOrderItemParams{
				ID:           item.ID(),
				OrderID:      item.OrderID(),
				TicketTypeID: item.TicketTypeID(),
				Quantity:     item.Quantity(),
				UnitPrice:    item.UnitPrice(),
				Price:        item.Price(),
				Currency:     item.Currency(),
			}); err != nil {
				return fmt.Errorf("insert order item: %w", err)
			}
		}

		domainEvents := order.DomainEvents()
		order.ClearDomainEvents()
		return outbox.InsertMessages(ctx, tx, schema, domainEvents)
	})
}

func (r *OrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	row, err := r.queries.GetOrderByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("get order: %w", err)
	}
	return domain.RehydrateOrder(
		row.ID,
		row.CustomerID,
		domain.OrderStatus(row.Status),
		row.TotalPrice,
		row.Currency,
		row.TicketsIssued,
		row.CreatedAtUtc.Time,
		nil,
	), nil
}
