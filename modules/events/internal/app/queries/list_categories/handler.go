package listcategories

import (
	"context"
	"fmt"
)

type CategoryReader interface {
	ListCategories(ctx context.Context) ([]CategoryItem, error)
}

type Handler struct {
	reader CategoryReader
}

func NewHandler(reader CategoryReader) *Handler {
	return &Handler{reader: reader}
}

func (h *Handler) Handle(ctx context.Context) ([]CategoryItem, error) {
	items, err := h.reader.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return items, nil
}
