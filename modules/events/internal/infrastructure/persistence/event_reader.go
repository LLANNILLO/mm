package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	getevent "github.com/llannillo/mm/modules/events/internal/application/queries/get-event"
	listevents "github.com/llannillo/mm/modules/events/internal/application/queries/list_events"
	searchevents "github.com/llannillo/mm/modules/events/internal/application/queries/search_events"
	"github.com/llannillo/mm/modules/events/internal/domain"
	store "github.com/llannillo/mm/modules/events/internal/infrastructure/store/generated"
)

type EventReader struct {
	queries *store.Queries
}

func NewEventReader(q *store.Queries) *EventReader {
	return &EventReader{queries: q}
}

func (r *EventReader) GetEvent(ctx context.Context, id uuid.UUID) (*getevent.Response, error) {
	row, err := r.queries.SelectEventByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEventNotFound
		}
		return nil, fmt.Errorf("get event: %w", err)
	}

	ttRows, err := r.queries.SelectTicketTypesByEventID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get ticket types: %w", err)
	}

	resp := &getevent.Response{
		ID:          row.ID,
		CategoryID:  row.CategoryID,
		Title:       row.Title,
		Description: row.Description,
		Location:    row.Location,
		StartsAtUtc: row.StartsAtUtc,
		EndsAtUtc:   row.EndsAtUtc,
		TicketTypes: make([]getevent.TicketTypeItem, len(ttRows)),
	}
	for i, tt := range ttRows {
		resp.TicketTypes[i] = getevent.TicketTypeItem{
			ID:       tt.ID,
			Name:     tt.Name,
			Price:    tt.Price,
			Currency: tt.Currency,
			Quantity: tt.Quantity,
		}
	}
	return resp, nil
}

func (r *EventReader) ListEvents(ctx context.Context) ([]listevents.EventItem, error) {
	rows, err := r.queries.SelectEvents(ctx)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	items := make([]listevents.EventItem, len(rows))
	for i, row := range rows {
		items[i] = listevents.EventItem{
			ID:          row.ID,
			CategoryID:  row.CategoryID,
			Title:       row.Title,
			StartsAtUtc: row.StartsAtUtc,
			EndsAtUtc:   row.EndsAtUtc,
		}
	}
	return items, nil
}

func (r *EventReader) SearchEvents(ctx context.Context, q searchevents.Query) ([]searchevents.EventItem, int64, error) {
	offset := (q.Page - 1) * q.PageSize

	rows, err := r.queries.SSearchEvents(ctx, store.SSearchEventsParams{
		Status:      q.Status,
		CategoryID:  q.CategoryID,
		StartsDate:  q.StartsFrom,
		EndsDate:    q.EndsFrom,
		OffsetCount: offset,
		LimitCount:  q.PageSize,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("search events: %w", err)
	}

	total, err := r.queries.SCountSearchEvents(ctx, store.SCountSearchEventsParams{
		Status:     q.Status,
		CategoryID: q.CategoryID,
		StartsDate: q.StartsFrom,
		EndsDate:   q.EndsFrom,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("count search events: %w", err)
	}

	items := make([]searchevents.EventItem, len(rows))
	for i, row := range rows {
		items[i] = searchevents.EventItem{
			ID:          row.ID,
			CategoryID:  row.CategoryID,
			Title:       row.Title,
			StartsAtUtc: row.StartsAtUtc,
			EndsAtUtc:   row.EndsAtUtc,
		}
	}
	return items, total, nil
}
