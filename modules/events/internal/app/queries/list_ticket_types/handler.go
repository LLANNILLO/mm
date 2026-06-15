package listtickettype

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type TicketTypeReader interface {
	ListTicketTypes(ctx context.Context, eventID uuid.UUID) ([]TicketTypeItem, error)
}

type Handler struct {
	reader TicketTypeReader
}

func NewHandler(reader TicketTypeReader) *Handler {
	return &Handler{reader: reader}
}

func (h *Handler) Handle(ctx context.Context, q Query) ([]TicketTypeItem, error) {
	items, err := h.reader.ListTicketTypes(ctx, q.EventID)
	if err != nil {
		return nil, fmt.Errorf("list ticket types: %w", err)
	}
	return items, nil
}
