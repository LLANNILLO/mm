package createtickettype

import (
	"context"
	"fmt"

	"github.com/llannillo/mm/modules/ticketing/internal/domain"
	"github.com/llannillo/mm/modules/ticketing/internal/ports/outbound"
)

type Handler struct {
	ticketTypeRepo outbound.TicketTypeRepository
}

func NewHandler(ticketTypeRepo outbound.TicketTypeRepository) *Handler {
	return &Handler{ticketTypeRepo: ticketTypeRepo}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	ticketType := domain.NewTicketType(cmd.ID, cmd.EventID, cmd.Name, cmd.Price, cmd.Currency, cmd.Quantity)

	if err := h.ticketTypeRepo.Insert(ctx, ticketType); err != nil {
		return fmt.Errorf("create ticket type: %w", err)
	}

	return nil
}
