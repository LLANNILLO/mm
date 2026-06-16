package publishevent

import (
	"context"
	"fmt"

	"github.com/llannillo/mm/modules/events/internal/domain"
	"github.com/llannillo/mm/modules/events/internal/ports/outbound"
)

type Handler struct {
	repo           outbound.EventRepository
	ticketTypeRepo outbound.TicketTypeRepository
}

func NewHandler(repo outbound.EventRepository, ticketTypeRepo outbound.TicketTypeRepository) *Handler {
	return &Handler{repo: repo, ticketTypeRepo: ticketTypeRepo}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	if err := cmd.Validate(); err != nil {
		return err
	}
	event, err := h.repo.GetByID(ctx, cmd.EventID)
	if err != nil {
		return err
	}
	hasTickets, err := h.ticketTypeRepo.ExistsByEventID(ctx, cmd.EventID)
	if err != nil {
		return fmt.Errorf("check ticket types: %w", err)
	}
	if !hasTickets {
		return domain.ErrEventHasNoTickets
	}
	if err := event.Publish(); err != nil {
		return fmt.Errorf("publish event: %w", err)
	}
	return h.repo.Update(ctx, event)
}
