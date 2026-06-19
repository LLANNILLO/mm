package createorder

import (
	"context"
	"fmt"

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

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	order := domain.NewOrder(cmd.CustomerID)

	for _, item := range cmd.TicketTypes {
		ticketType, err := h.ticketTypeRepo.GetByID(ctx, item.TicketTypeID)
		if err != nil {
			return fmt.Errorf("get ticket type: %w", err)
		}

		if err := ticketType.UpdateQuantity(item.Quantity); err != nil {
			return fmt.Errorf("update ticket type quantity: %w", err)
		}

		if err := order.AddItem(ticketType, item.Quantity); err != nil {
			return fmt.Errorf("add item to order: %w", err)
		}
	}

	if err := h.orderRepo.Create(ctx, order); err != nil {
		return fmt.Errorf("create order: %w", err)
	}

	return nil
}
