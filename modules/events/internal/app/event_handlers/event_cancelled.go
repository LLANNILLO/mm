package eventhandlers

import (
	"context"

	"github.com/llannillo/mm/internal/shared/eventbus"
	"github.com/llannillo/mm/modules/events/api/integrationevents"
	"github.com/llannillo/mm/modules/events/internal/domain"
)

// EventCancelledHandler republishes the local EventCancelledDomainEvent as
// the module's public EventCanceledIntegrationEvent, so other modules (and
// the cancel-event saga) can react without depending on this module's
// internals.
type EventCancelledHandler struct {
	eventBus eventbus.EventBus
}

func NewEventCancelledHandler(eventBus eventbus.EventBus) *EventCancelledHandler {
	return &EventCancelledHandler{eventBus: eventBus}
}

func (h *EventCancelledHandler) Handle(ctx context.Context, e domain.EventCancelledDomainEvent) error {
	return h.eventBus.Publish(ctx, integrationevents.EventCanceledIntegrationEvent{EventID: e.EventID})
}
