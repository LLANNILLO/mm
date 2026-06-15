package searchevents

import (
	"context"
	"fmt"
)

type EventsReader interface {
	SearchEvents(ctx context.Context, q Query) ([]EventItem, int64, error)
}

type Handler struct {
	reader EventsReader
}

func NewHandler(reader EventsReader) *Handler {
	return &Handler{reader: reader}
}

func (h *Handler) Handle(ctx context.Context, q Query) (*Page[EventItem], error) {
	items, total, err := h.reader.SearchEvents(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("search events: %w", err)
	}
	return &Page[EventItem]{
		Items:      items,
		TotalCount: total,
		Page:       q.Page,
		PageSize:   q.PageSize,
	}, nil
}
