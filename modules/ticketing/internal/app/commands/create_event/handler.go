package createevent

import (
	"context"
	"fmt"

	"github.com/llannillo/mm/modules/ticketing/internal/domain"
	"github.com/llannillo/mm/modules/ticketing/internal/ports/outbound"
)

type Handler struct {
	eventRepo      outbound.EventRepository
	ticketTypeRepo outbound.TicketTypeRepository
}

func NewHandler(eventRepo outbound.EventRepository, ticketTypeRepo outbound.TicketTypeRepository) *Handler {
	return &Handler{
		eventRepo:      eventRepo,
		ticketTypeRepo: ticketTypeRepo,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	event := domain.NewEvent(
		cmd.EventID,
		cmd.Title,
		cmd.Description,
		cmd.Location,
		cmd.StartsAtUtc,
		cmd.EndsAtUtc,
	)

	if err := h.eventRepo.Insert(ctx, event); err != nil {
		return fmt.Errorf("create event: %w", err)
	}

	ticketTypes := make([]*domain.TicketType, 0, len(cmd.TicketTypes))
	for _, tt := range cmd.TicketTypes {
		ticketTypes = append(ticketTypes, domain.NewTicketType(
			tt.ID,
			tt.EventID,
			tt.Name,
			tt.Price,
			tt.Currency,
			tt.Quantity,
		))
	}

	if err := h.ticketTypeRepo.InsertBatch(ctx, ticketTypes); err != nil {
		return fmt.Errorf("create ticket types: %w", err)
	}

	return nil
}
