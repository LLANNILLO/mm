package rescheduleevent

import (
	"context"
	"fmt"

	"github.com/llannillo/mm/modules/events/internal/ports/outbound"
)

type Handler struct {
	repo outbound.EventRepository
}

func NewHandler(repo outbound.EventRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	event, err := h.repo.GetByID(ctx, cmd.EventID)
	if err != nil {
		return err
	}
	if err := event.Reschedule(cmd.StartsAtUtc, cmd.EndsAtUtc); err != nil {
		return fmt.Errorf("reschedule event: %w", err)
	}
	return h.repo.Update(ctx, event)
}
