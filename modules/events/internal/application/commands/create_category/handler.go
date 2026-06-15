package createcategory

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/llannillo/mm/modules/events/internal/domain"
)

type Handler struct {
	repo domain.CategoryRepository
}

func NewHandler(repo domain.CategoryRepository) *Handler {
	return &Handler{
		repo: repo,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (uuid.UUID, error) {
	category, err := domain.NewCategory(cmd.Name)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create category: %w", err)
	}
	if err := h.repo.Insert(ctx, category); err != nil {
		return uuid.Nil, err
	}
	return category.ID, nil
}
