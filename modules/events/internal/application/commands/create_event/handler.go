package createevent

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/llannillo/mm/modules/events/internal/domain"
)

type Handler struct {
	repo  domain.EventRepository
	clock domain.Clock
}

func NewHandler(repo domain.EventRepository, clock domain.Clock) *Handler {
	return &Handler{repo: repo, clock: clock}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (uuid.UUID, error) {
	event, err := domain.NewEvent(
		cmd.CategoryID,
		cmd.Title,
		cmd.Description,
		cmd.Location,
		cmd.StartsAtUtc,
		cmd.EndsAtUtc,
		h.clock.Now(),
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create event: %w", err)
	}
	if err := h.repo.Insert(ctx, event); err != nil {
		return uuid.Nil, err
	}
	return event.ID, nil
}
