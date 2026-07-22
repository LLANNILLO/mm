package eventhandlers

import (
	"context"
	"fmt"

	"github.com/llannillo/mm/internal/shared/eventbus"
	"github.com/llannillo/mm/modules/events/api/integrationevents"
	gettickettype "github.com/llannillo/mm/modules/events/internal/app/queries/get_ticket_type"
	"github.com/llannillo/mm/modules/events/internal/domain"
)

// TicketTypeCreatedHandler republishes the local TicketTypeCreatedDomainEvent
// as the module's public TicketTypeCreatedIntegrationEvent, re-querying the
// full ticket type first since the domain event only carries its ID.
type TicketTypeCreatedHandler struct {
	getTicketType *gettickettype.Handler
	eventBus      eventbus.EventBus
}

func NewTicketTypeCreatedHandler(getTicketType *gettickettype.Handler, eventBus eventbus.EventBus) *TicketTypeCreatedHandler {
	return &TicketTypeCreatedHandler{getTicketType: getTicketType, eventBus: eventBus}
}

func (h *TicketTypeCreatedHandler) Handle(ctx context.Context, e domain.TicketTypeCreatedDomainEvent) error {
	resp, err := h.getTicketType.Handle(ctx, gettickettype.Query{ID: e.TicketTypeID})
	if err != nil {
		return fmt.Errorf("get ticket type: %w", err)
	}
	return h.eventBus.Publish(ctx, integrationevents.TicketTypeCreatedIntegrationEvent{
		TicketTypeID: resp.ID,
		EventID:      resp.EventID,
		Name:         resp.Name,
		Price:        resp.Price,
		Currency:     resp.Currency,
		Quantity:     resp.Quantity,
	})
}
