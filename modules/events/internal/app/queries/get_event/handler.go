package getevent

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type EventReader interface {
	GetEvent(ctx context.Context, id uuid.UUID) (*Response, error)
}

type Handler struct {
	reader EventReader
}

func NewHandler(reader EventReader) *Handler {
	return &Handler{reader: reader}
}

func (h *Handler) Handle(ctx context.Context, q Query) (*Response, error) {
	resp, err := h.reader.GetEvent(ctx, q.ID)
	if err != nil {
		return nil, fmt.Errorf("get event: %w", err)
	}
	return resp, nil
}
