package checkinticket

import (
	"context"
	"fmt"

	"github.com/llannillo/mm/modules/ticketing/internal/ports/outbound"
)

type Handler struct {
	ticketRepo outbound.TicketRepository
}

func NewHandler(ticketRepo outbound.TicketRepository) *Handler {
	return &Handler{ticketRepo: ticketRepo}
}

// Handle checks a ticket in. It persists the ticket even when CheckIn fails:
// the invalid/duplicate-attempt domain event still needs to reach the
// outbox so the event_statistics projections can count it.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	ticket, err := h.ticketRepo.GetByID(ctx, cmd.TicketID)
	if err != nil {
		return fmt.Errorf("get ticket: %w", err)
	}

	checkInErr := ticket.CheckIn(cmd.CustomerID)

	if err := h.ticketRepo.Update(ctx, ticket); err != nil {
		return fmt.Errorf("update ticket: %w", err)
	}

	return checkInErr
}
