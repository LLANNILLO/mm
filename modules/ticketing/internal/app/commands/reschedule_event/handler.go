package rescheduleevent

import (
	"context"
	"fmt"

	"github.com/llannillo/mm/modules/ticketing/internal/ports/outbound"
)

type Handler struct {
	eventRepo outbound.EventRepository
}

func NewHandler(eventRepo outbound.EventRepository) *Handler {
	return &Handler{eventRepo: eventRepo}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	event, err := h.eventRepo.GetByID(ctx, cmd.EventID)
	if err != nil {
		return fmt.Errorf("get event: %w", err)
	}

	event.Reschedule(cmd.StartsAtUtc, cmd.EndsAtUtc)

	if err := h.eventRepo.Update(ctx, event); err != nil {
		return fmt.Errorf("reschedule event: %w", err)
	}

	return nil
}
