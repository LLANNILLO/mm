package renamecategory

import (
	"context"
	"fmt"

	"github.com/llannillo/mm/modules/events/internal/ports/outbound"
)

type Handler struct {
	repo outbound.CategoryRepository
}

func NewHandler(repo outbound.CategoryRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	if err := cmd.Validate(); err != nil {
		return err
	}
	category, err := h.repo.GetByID(ctx, cmd.CategoryID)
	if err != nil {
		return err
	}
	if err := category.ChangeName(cmd.Name); err != nil {
		return fmt.Errorf("rename category: %w", err)
	}
	return h.repo.Update(ctx, category)
}
