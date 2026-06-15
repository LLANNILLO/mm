package getcategory

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type CategoryReader interface {
	GetCategory(ctx context.Context, id uuid.UUID) (*Response, error)
}

type Handler struct {
	reader CategoryReader
}

func NewHandler(reader CategoryReader) *Handler {
	return &Handler{reader: reader}
}

func (h *Handler) Handle(ctx context.Context, q Query) (*Response, error) {
	resp, err := h.reader.GetCategory(ctx, q.ID)
	if err != nil {
		return nil, fmt.Errorf("get category: %w", err)
	}
	return resp, nil
}
