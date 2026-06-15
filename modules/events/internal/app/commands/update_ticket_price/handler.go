package updateticketprice

import (
	"context"
	"fmt"

	"github.com/llannillo/mm/modules/events/internal/ports/outbound"
)

type Handler struct {
	repo outbound.TicketTypeRepository
}

func NewHandler(repo outbound.TicketTypeRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	tt, err := h.repo.GetByID(ctx, cmd.TicketTypeID)
	if err != nil {
		return err
	}
	if err := tt.UpdatePrice(cmd.Price); err != nil {
		return fmt.Errorf("update ticket price: %w", err)
	}
	return h.repo.Update(ctx, tt)
}
