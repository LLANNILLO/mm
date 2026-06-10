package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/llannillo/mm/modules/events/internal/domain"
	"github.com/llannillo/mm/modules/events/internal/store"
)

type EventRepository struct {
	queries *store.Queries
}

func NewEventRepository(q *store.Queries) *EventRepository {
	return &EventRepository{queries: q}
}

func (r *EventRepository) Insert(ctx context.Context, event *domain.Event) error {
	endsAt := time.Time{}
	if event.EndsAtUtc != nil {
		endsAt = *event.EndsAtUtc
	}
	_, err := r.queries.CreateEvent(ctx, store.CreateEventParams{
		ID:          event.ID,
		Title:       event.Title,
		Description: event.Description,
		Location:    event.Location,
		StartsAtUtc: event.StartsAtUtc,
		EndsAtUtc:   endsAt,
		Status:      event.Status,
	})
	if err != nil {
		return fmt.Errorf("insert event: %w", err)
	}
	return nil
}
