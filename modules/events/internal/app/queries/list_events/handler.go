package listevents

import (
	"context"
	"fmt"
)

type EventsReader interface {
	ListEvents(ctx context.Context) ([]EventItem, error)
}

type Handler struct {
	reader EventsReader
}

func NewHandler(reader EventsReader) *Handler {
	return &Handler{reader: reader}
}

func (h *Handler) Handle(ctx context.Context) ([]EventItem, error) {
	items, err := h.reader.ListEvents(ctx)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	return items, nil
}
