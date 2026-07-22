package eventhandlers

import (
	"context"
	"fmt"

	"github.com/llannillo/mm/internal/shared/eventbus"
	"github.com/llannillo/mm/modules/events/api/integrationevents"
	getevent "github.com/llannillo/mm/modules/events/internal/app/queries/get_event"
	"github.com/llannillo/mm/modules/events/internal/domain"
)

// EventCreatedHandler republishes the local EventCreatedDomainEvent as the
// module's public EventCreatedIntegrationEvent. The domain event only carries
// the new event's ID, so this re-queries the full event before publishing —
// same pattern as UserRegisteredHandler in the users module.
type EventCreatedHandler struct {
	getEvent *getevent.Handler
	eventBus eventbus.EventBus
}

func NewEventCreatedHandler(getEvent *getevent.Handler, eventBus eventbus.EventBus) *EventCreatedHandler {
	return &EventCreatedHandler{getEvent: getEvent, eventBus: eventBus}
}

func (h *EventCreatedHandler) Handle(ctx context.Context, e domain.EventCreatedDomainEvent) error {
	resp, err := h.getEvent.Handle(ctx, getevent.Query{ID: e.EventID})
	if err != nil {
		return fmt.Errorf("get event: %w", err)
	}
	return h.eventBus.Publish(ctx, integrationevents.EventCreatedIntegrationEvent{
		EventID:     resp.ID,
		CategoryID:  resp.CategoryID,
		Title:       resp.Title,
		Description: resp.Description,
		Location:    resp.Location,
		StartsAtUtc: resp.StartsAtUtc,
		EndsAtUtc:   resp.EndsAtUtc,
	})
}
