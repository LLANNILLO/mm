package eventhandlers

import (
	"context"

	"github.com/llannillo/mm/internal/shared/eventbus"
	"github.com/llannillo/mm/modules/events/api/integrationevents"
	"github.com/llannillo/mm/modules/events/internal/domain"
)

// TicketTypePriceChangedHandler republishes the local
// TicketTypePriceChangedDomainEvent as the module's public
// TicketTypePriceChangedIntegrationEvent. The domain event already carries
// the new price, so no re-query is needed.
type TicketTypePriceChangedHandler struct {
	eventBus eventbus.EventBus
}

func NewTicketTypePriceChangedHandler(eventBus eventbus.EventBus) *TicketTypePriceChangedHandler {
	return &TicketTypePriceChangedHandler{eventBus: eventBus}
}

func (h *TicketTypePriceChangedHandler) Handle(ctx context.Context, e domain.TicketTypePriceChangedDomainEvent) error {
	return h.eventBus.Publish(ctx, integrationevents.TicketTypePriceChangedIntegrationEvent{
		TicketTypeID: e.TicketTypeID,
		Price:        e.Price,
	})
}
