package createevent

import (
	"context"

	"github.com/google/uuid"
	"github.com/llannillo/mm/modules/events/internal/domain"
)

type Handler struct {
	repo domain.EventRepository
}

func NewHandler(repo domain.EventRepository) *Handler {
	return &Handler{
		repo: repo,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (uuid.UUID, error) {
	event := domain.NewEvent(
		cmd.Title,
		cmd.Description,
		cmd.Location,
		cmd.StartsAtUtc,
		cmd.EndsAtUtc,
	)
	if err := h.repo.Insert(ctx, event); err != nil {
		return uuid.Nil, err
	}
	return event.ID, nil
}
