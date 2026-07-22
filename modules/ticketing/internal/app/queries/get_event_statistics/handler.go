package geteventstatistics

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type EventStatisticsReader interface {
	GetEventStatistics(ctx context.Context, eventID uuid.UUID) (*Response, error)
}

type Handler struct {
	reader EventStatisticsReader
}

func NewHandler(reader EventStatisticsReader) *Handler {
	return &Handler{reader: reader}
}

func (h *Handler) Handle(ctx context.Context, q Query) (*Response, error) {
	resp, err := h.reader.GetEventStatistics(ctx, q.EventID)
	if err != nil {
		return nil, fmt.Errorf("get event statistics: %w", err)
	}
	return resp, nil
}
