package updatecustomer

import (
	"context"
	"fmt"

	"github.com/llannillo/mm/modules/ticketing/internal/ports/outbound"
)

type Handler struct {
	repo outbound.CustomerRepository
}

func NewHandler(repo outbound.CustomerRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	customer, err := h.repo.GetByID(ctx, cmd.ID)
	if err != nil {
		return fmt.Errorf("get customer: %w", err)
	}
	customer.Update(cmd.FirstName, cmd.LastName)
	if err := h.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("update customer: %w", err)
	}
	return nil
}
