package renamecategory

import (
	"context"
	"fmt"

	"github.com/llannillo/mm/modules/events/internal/domain"
)

type Handler struct {
	repo domain.CategoryRepository
}

func NewHandler(repo domain.CategoryRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	category, err := h.repo.GetByID(ctx, cmd.CategoryID)
	if err != nil {
		return err
	}
	if err := category.ChangeName(cmd.Name); err != nil {
		return fmt.Errorf("rename category: %w", err)
	}
	return h.repo.Update(ctx, category)
}
