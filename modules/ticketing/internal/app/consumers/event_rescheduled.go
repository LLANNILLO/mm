package consumers

import (
	"context"

	eventsintegrationevents "github.com/llannillo/mm/modules/events/api/integrationevents"
	rescheduleevent "github.com/llannillo/mm/modules/ticketing/internal/app/commands/reschedule_event"
)

type EventRescheduledConsumer struct {
	rescheduleEvent *rescheduleevent.Handler
}

func NewEventRescheduledConsumer(h *rescheduleevent.Handler) *EventRescheduledConsumer {
	return &EventRescheduledConsumer{rescheduleEvent: h}
}

func (c *EventRescheduledConsumer) Handle(ctx context.Context, e eventsintegrationevents.EventRescheduledIntegrationEvent) error {
	return c.rescheduleEvent.Handle(ctx, rescheduleevent.Command{
		EventID:     e.EventID,
		StartsAtUtc: e.StartsAtUtc,
		EndsAtUtc:   e.EndsAtUtc,
	})
}
