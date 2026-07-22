package eventhandlers

import (
	"context"

	"github.com/llannillo/mm/internal/shared/eventbus"
	"github.com/llannillo/mm/modules/events/api/integrationevents"
	"github.com/llannillo/mm/modules/events/internal/domain"
)

// EventRescheduledHandler republishes the local EventRescheduledDomainEvent as
// the module's public EventRescheduledIntegrationEvent. Unlike creation, the
// domain event already carries the new schedule, so no re-query is needed.
type EventRescheduledHandler struct {
	eventBus eventbus.EventBus
}

func NewEventRescheduledHandler(eventBus eventbus.EventBus) *EventRescheduledHandler {
	return &EventRescheduledHandler{eventBus: eventBus}
}

func (h *EventRescheduledHandler) Handle(ctx context.Context, e domain.EventRescheduledDomainEvent) error {
	return h.eventBus.Publish(ctx, integrationevents.EventRescheduledIntegrationEvent{
		EventID:     e.EventID,
		StartsAtUtc: e.StartsAtUtc,
		EndsAtUtc:   e.EndsAtUtc,
	})
}
