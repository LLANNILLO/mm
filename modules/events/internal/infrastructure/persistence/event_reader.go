package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	getevent "github.com/llannillo/mm/modules/events/internal/application/queries/get-event"
	"github.com/llannillo/mm/modules/events/internal/store"
)

type EventReader struct {
	queries *store.Queries
}

func NewEventReader(q *store.Queries) *EventReader {
	return &EventReader{queries: q}
}

func (r *EventReader) GetEvent(ctx context.Context, id uuid.UUID) (*getevent.Response, error) {
	row, err := r.queries.GetEvent(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get event row: %w", err)
	}
	return toResponse(row), nil
}

func toResponse(row store.GetEventRow) *getevent.Response {
	resp := &getevent.Response{
		ID:          row.ID,
		Title:       row.Title,
		Description: row.Description,
		Location:    row.Location,
		StartsAtUtc: row.StartsAtUtc,
	}
	if !row.EndsAtUtc.IsZero() {
		t := row.EndsAtUtc
		resp.EndsAtUtc = &t
	}
	return resp
}
