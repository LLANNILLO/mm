package getuser

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type UserReader interface {
	GetUser(ctx context.Context, id uuid.UUID) (*Response, error)
}

type Handler struct {
	reader UserReader
}

func NewHandler(reader UserReader) *Handler {
	return &Handler{reader: reader}
}

func (h *Handler) Handle(ctx context.Context, q Query) (*Response, error) {
	resp, err := h.reader.GetUser(ctx, q.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return resp, nil
}
