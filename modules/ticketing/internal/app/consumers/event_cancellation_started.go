package consumers

import (
	"context"

	cancelevent "github.com/llannillo/mm/modules/ticketing/internal/app/commands/cancel_event"

	eventsintegrationevents "github.com/llannillo/mm/modules/events/api/integrationevents"
)

type EventCancellationStartedConsumer struct {
	cancelEvent *cancelevent.Handler
}

func NewEventCancellationStartedConsumer(h *cancelevent.Handler) *EventCancellationStartedConsumer {
	return &EventCancellationStartedConsumer{cancelEvent: h}
}

func (c *EventCancellationStartedConsumer) Handle(ctx context.Context, e eventsintegrationevents.EventCancellationStartedIntegrationEvent) error {
	return c.cancelEvent.Handle(ctx, cancelevent.Command{EventID: e.EventID})
}
