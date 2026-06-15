package cancelevent

import (
	"context"
	"fmt"

	"github.com/llannillo/mm/modules/events/internal/domain"
	"github.com/llannillo/mm/modules/events/internal/ports/outbound"
)

type Handler struct {
	repo  outbound.EventRepository
	clock domain.Clock
}

func NewHandler(repo outbound.EventRepository, clock domain.Clock) *Handler {
	return &Handler{repo: repo, clock: clock}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	event, err := h.repo.GetByID(ctx, cmd.EventID)
	if err != nil {
		return err
	}
	if err := event.Cancel(h.clock.Now()); err != nil {
		return fmt.Errorf("cancel event: %w", err)
	}
	return h.repo.Update(ctx, event)
}
