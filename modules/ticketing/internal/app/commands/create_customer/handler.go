package createcustomer

import (
	"context"
	"fmt"

	"github.com/llannillo/mm/modules/ticketing/internal/domain"
	"github.com/llannillo/mm/modules/ticketing/internal/ports/outbound"
)

type Handler struct {
	repo outbound.CustomerRepository
}

func NewHandler(repo outbound.CustomerRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	customer, err := domain.NewCustomer(cmd.ID, cmd.Email, cmd.FirstName, cmd.LastName)
	if err != nil {
		return err
	}
	if err := h.repo.Insert(ctx, customer); err != nil {
		return fmt.Errorf("create customer: %w", err)
	}
	return nil
}
