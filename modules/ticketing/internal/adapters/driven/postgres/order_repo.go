package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgtype"
	store "github.com/llannillo/mm/modules/ticketing/internal/adapters/driven/postgres/generated"
	"github.com/llannillo/mm/modules/ticketing/internal/domain"
)

type OrderRepository struct {
	pool    *pgxpool.Pool
	queries *store.Queries
}

func NewOrderRepository(pool *pgxpool.Pool, q *store.Queries) *OrderRepository {
	return &OrderRepository{pool: pool, queries: q}
}

func (r *OrderRepository) Create(ctx context.Context, order *domain.Order) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	q := r.queries.WithTx(tx)

	// For each order item: lock the ticket_type row, validate and decrement available_quantity
	for _, item := range order.Items {
		var availableQty int64
		row := tx.QueryRow(ctx,
			`SELECT available_quantity FROM ticketing.ticket_types WHERE id = $1 FOR UPDATE`,
			item.TicketTypeID,
		)
		if err := row.Scan(&availableQty); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrTicketTypeNotFound
			}
			return fmt.Errorf("lock ticket type: %w", err)
		}

		if item.Quantity > availableQty {
			return domain.ErrTicketTypeInsufficientQuantity
		}

		if err := q.DecrementTicketTypeQuantity(ctx, store.DecrementTicketTypeQuantityParams{
			AvailableQuantity: item.Quantity,
			ID:                item.TicketTypeID,
		}); err != nil {
			return fmt.Errorf("decrement ticket type quantity: %w", err)
		}
	}

	// Insert the order
	createdAtUtc := pgtype.Timestamptz{Time: order.CreatedAtUtc, Valid: true}
	if err := q.InsertOrder(ctx, store.InsertOrderParams{
		ID:            order.ID,
		CustomerID:    order.CustomerID,
		Status:        string(order.Status),
		TotalPrice:    order.TotalPrice,
		Currency:      order.Currency,
		TicketsIssued: order.TicketsIssued,
		CreatedAtUtc:  createdAtUtc,
	}); err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	// Insert all order items
	for _, item := range order.Items {
		if err := q.InsertOrderItem(ctx, store.InsertOrderItemParams{
			ID:           item.ID,
			OrderID:      item.OrderID,
			TicketTypeID: item.TicketTypeID,
			Quantity:     item.Quantity,
			UnitPrice:    item.UnitPrice,
			Price:        item.Price,
			Currency:     item.Currency,
		}); err != nil {
			return fmt.Errorf("insert order item: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	order.ClearDomainEvents()
	return nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	row, err := r.queries.GetOrderByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("get order: %w", err)
	}
	return &domain.Order{
		ID:            row.ID,
		CustomerID:    row.CustomerID,
		Status:        domain.OrderStatus(row.Status),
		TotalPrice:    row.TotalPrice,
		Currency:      row.Currency,
		TicketsIssued: row.TicketsIssued,
		CreatedAtUtc:  row.CreatedAtUtc.Time,
	}, nil
}
