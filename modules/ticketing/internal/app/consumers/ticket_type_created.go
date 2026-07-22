package consumers

import (
	"context"

	eventsintegrationevents "github.com/llannillo/mm/modules/events/api/integrationevents"
	createtickettype "github.com/llannillo/mm/modules/ticketing/internal/app/commands/create_ticket_type"
)

type TicketTypeCreatedConsumer struct {
	createTicketType *createtickettype.Handler
}

func NewTicketTypeCreatedConsumer(h *createtickettype.Handler) *TicketTypeCreatedConsumer {
	return &TicketTypeCreatedConsumer{createTicketType: h}
}

func (c *TicketTypeCreatedConsumer) Handle(ctx context.Context, e eventsintegrationevents.TicketTypeCreatedIntegrationEvent) error {
	return c.createTicketType.Handle(ctx, createtickettype.Command{
		ID:       e.TicketTypeID,
		EventID:  e.EventID,
		Name:     e.Name,
		Price:    e.Price,
		Currency: e.Currency,
		Quantity: e.Quantity,
	})
}
