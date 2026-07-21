package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/llannillo/mm/internal/shared/events"
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

		var domainEvents []events.DomainEvent

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

			remaining, err := q.DecrementTicketTypeQuantity(ctx, store.DecrementTicketTypeQuantityParams{
				AvailableQuantity: item.Quantity(),
				ID:                item.TicketTypeID(),
			})
			if err != nil {
				return fmt.Errorf("decrement ticket type quantity: %w", err)
			}
			if remaining == 0 {
				domainEvents = append(domainEvents, domain.TicketTypeSoldOutDomainEvent{TicketTypeID: item.TicketTypeID()})
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

		domainEvents = append(domainEvents, order.DomainEvents()...)
		order.ClearDomainEvents()
		return outbox.InsertMessages(ctx, tx, schema, domainEvents)
	})
}

func (r *OrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	row, err := r.queries.GetOrderByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("get order: %w", err)
	}

	itemRows, err := r.queries.GetOrderItemsByOrderID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get order items: %w", err)
	}
	items := make([]domain.OrderItem, 0, len(itemRows))
	for _, ir := range itemRows {
		items = append(items, domain.RehydrateOrderItem(ir.ID, ir.OrderID, ir.TicketTypeID, ir.Quantity, ir.UnitPrice, ir.Price, ir.Currency))
	}

	return domain.RehydrateOrder(
		row.ID,
		row.CustomerID,
		domain.OrderStatus(row.Status),
		row.TotalPrice,
		row.Currency,
		row.TicketsIssued,
		row.CreatedAtUtc.Time,
		items,
	), nil
}

// IssueTickets creates one Ticket per unit of quantity across the order's
// items and marks the order as having its tickets issued — atomically, in a
// single transaction. Called by the OrderCreated domain event handler.
func (r *OrderRepository) IssueTickets(ctx context.Context, orderID uuid.UUID) error {
	return r.uow.WithTx(ctx, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		orderRow, err := q.GetOrderByID(ctx, orderID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrOrderNotFound
			}
			return fmt.Errorf("get order: %w", err)
		}

		itemRows, err := q.GetOrderItemsByOrderID(ctx, orderID)
		if err != nil {
			return fmt.Errorf("get order items: %w", err)
		}
		items := make([]domain.OrderItem, 0, len(itemRows))
		for _, ir := range itemRows {
			items = append(items, domain.RehydrateOrderItem(ir.ID, ir.OrderID, ir.TicketTypeID, ir.Quantity, ir.UnitPrice, ir.Price, ir.Currency))
		}

		order := domain.RehydrateOrder(
			orderRow.ID, orderRow.CustomerID, domain.OrderStatus(orderRow.Status),
			orderRow.TotalPrice, orderRow.Currency, orderRow.TicketsIssued, orderRow.CreatedAtUtc.Time, items,
		)

		if err := order.IssueTickets(); err != nil {
			if errors.Is(err, domain.ErrOrderTicketsAlreadyIssued) {
				return nil
			}
			return err
		}

		var tickets []*domain.Ticket
		for _, item := range order.Items() {
			ttRow, err := q.GetTicketTypeByID(ctx, item.TicketTypeID())
			if err != nil {
				return fmt.Errorf("get ticket type %s: %w", item.TicketTypeID(), err)
			}
			ticketType := domain.RehydrateTicketType(
				ttRow.ID, ttRow.EventID, ttRow.Name, ttRow.Price, ttRow.Currency, ttRow.Quantity, ttRow.AvailableQuantity,
			)

			for i := int64(0); i < item.Quantity(); i++ {
				tickets = append(tickets, domain.NewTicket(order, ticketType))
			}
		}

		for _, t := range tickets {
			ticketCreatedAtUtc := pgtype.Timestamptz{Time: t.CreatedAtUtc(), Valid: true}
			if err := q.InsertTicket(ctx, store.InsertTicketParams{
				ID:           t.ID(),
				CustomerID:   t.CustomerID(),
				OrderID:      t.OrderID(),
				EventID:      t.EventID(),
				TicketTypeID: t.TicketTypeID(),
				Code:         t.Code(),
				CreatedAtUtc: ticketCreatedAtUtc,
				Archived:     t.Archived(),
			}); err != nil {
				return fmt.Errorf("insert ticket: %w", err)
			}
		}

		if err := q.UpdateOrderTicketsIssued(ctx, orderID); err != nil {
			return fmt.Errorf("mark tickets issued: %w", err)
		}

		domainEvents := append([]events.DomainEvent{}, order.DomainEvents()...)
		for _, t := range tickets {
			domainEvents = append(domainEvents, t.DomainEvents()...)
		}
		return outbox.InsertMessages(ctx, tx, schema, domainEvents)
	})
}
