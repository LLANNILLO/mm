package consumers

import (
	"context"

	eventsintegrationevents "github.com/llannillo/mm/modules/events/api/integrationevents"
	createevent "github.com/llannillo/mm/modules/ticketing/internal/app/commands/create_event"
)

type EventCreatedConsumer struct {
	createEvent *createevent.Handler
}

func NewEventCreatedConsumer(h *createevent.Handler) *EventCreatedConsumer {
	return &EventCreatedConsumer{createEvent: h}
}

func (c *EventCreatedConsumer) Handle(ctx context.Context, e eventsintegrationevents.EventCreatedIntegrationEvent) error {
	return c.createEvent.Handle(ctx, createevent.Command{
		EventID:     e.EventID,
		Title:       e.Title,
		Description: e.Description,
		Location:    e.Location,
		StartsAtUtc: e.StartsAtUtc,
		EndsAtUtc:   e.EndsAtUtc,
	})
}
