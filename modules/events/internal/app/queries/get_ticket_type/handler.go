package gettickettype

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type TicketTypeReader interface {
	GetTicketType(ctx context.Context, id uuid.UUID) (*Response, error)
}

type Handler struct {
	reader TicketTypeReader
}

func NewHandler(reader TicketTypeReader) *Handler {
	return &Handler{reader: reader}
}

func (h *Handler) Handle(ctx context.Context, q Query) (*Response, error) {
	resp, err := h.reader.GetTicketType(ctx, q.ID)
	if err != nil {
		return nil, fmt.Errorf("get ticket type: %w", err)
	}
	return resp, nil
}
