package cancelevent

import (
	"context"
	"fmt"

	"github.com/llannillo/mm/modules/events/internal/domain"
)

type Handler struct {
	repo  domain.EventRepository
	clock domain.Clock
}

func NewHandler(repo domain.EventRepository, clock domain.Clock) *Handler {
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
