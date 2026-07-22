package createevent

import (
	"context"
	"fmt"

	"github.com/llannillo/mm/modules/ticketing/internal/domain"
	"github.com/llannillo/mm/modules/ticketing/internal/ports/outbound"
)

type Handler struct {
	eventRepo outbound.EventRepository
}

func NewHandler(eventRepo outbound.EventRepository) *Handler {
	return &Handler{eventRepo: eventRepo}
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

	return nil
}
