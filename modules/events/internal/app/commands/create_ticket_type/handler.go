package createtickettype

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/llannillo/mm/modules/events/internal/domain"
	"github.com/llannillo/mm/modules/events/internal/ports/outbound"
)

type Handler struct {
	repo outbound.TicketTypeRepository
}

func NewHandler(repo outbound.TicketTypeRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (uuid.UUID, error) {
	if err := cmd.Validate(); err != nil {
		return uuid.Nil, err
	}
	tt, err := domain.NewTicketType(cmd.EventID, cmd.Name, cmd.Price, cmd.Currency, cmd.Quantity)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create ticket type: %w", err)
	}
	if err := h.repo.Insert(ctx, tt); err != nil {
		return uuid.Nil, err
	}
	return tt.ID(), nil
}
