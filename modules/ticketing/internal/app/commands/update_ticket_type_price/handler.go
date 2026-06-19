package updatetickettypeprice

import (
	"context"
	"fmt"

	"github.com/llannillo/mm/modules/ticketing/internal/ports/outbound"
)

type Handler struct {
	ticketTypeRepo outbound.TicketTypeRepository
}

func NewHandler(ticketTypeRepo outbound.TicketTypeRepository) *Handler {
	return &Handler{ticketTypeRepo: ticketTypeRepo}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	ticketType, err := h.ticketTypeRepo.GetByID(ctx, cmd.TicketTypeID)
	if err != nil {
		return fmt.Errorf("get ticket type: %w", err)
	}

	ticketType.UpdatePrice(cmd.Price)

	if err := h.ticketTypeRepo.Update(ctx, ticketType); err != nil {
		return fmt.Errorf("update ticket type price: %w", err)
	}

	return nil
}
