package createorder

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/llannillo/mm/modules/ticketing/internal/domain"
	"github.com/llannillo/mm/modules/ticketing/internal/ports/outbound"
)

type Handler struct {
	ticketTypeRepo outbound.TicketTypeRepository
	orderRepo      outbound.OrderRepository
}

func NewHandler(ticketTypeRepo outbound.TicketTypeRepository, orderRepo outbound.OrderRepository) *Handler {
	return &Handler{
		ticketTypeRepo: ticketTypeRepo,
		orderRepo:      orderRepo,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (uuid.UUID, error) {
	order := domain.NewOrder(cmd.CustomerID)

	// Availability is validated and decremented atomically, under a row lock,
	// inside orderRepo.Create — not here. TicketType is only read for its
	// price/currency to compute each order item's line total.
	for _, item := range cmd.TicketTypes {
		ticketType, err := h.ticketTypeRepo.GetByID(ctx, item.TicketTypeID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("get ticket type: %w", err)
		}

		if err := order.AddItem(ticketType, item.Quantity); err != nil {
			return uuid.Nil, fmt.Errorf("add item to order: %w", err)
		}
	}

	if err := h.orderRepo.Create(ctx, order); err != nil {
		return uuid.Nil, fmt.Errorf("create order: %w", err)
	}

	return order.ID(), nil
}
