package consumers

import (
	"context"

	eventsintegrationevents "github.com/llannillo/mm/modules/events/api/integrationevents"
	updatetickettypeprice "github.com/llannillo/mm/modules/ticketing/internal/app/commands/update_ticket_type_price"
)

type TicketTypePriceChangedConsumer struct {
	updateTicketTypePrice *updatetickettypeprice.Handler
}

func NewTicketTypePriceChangedConsumer(h *updatetickettypeprice.Handler) *TicketTypePriceChangedConsumer {
	return &TicketTypePriceChangedConsumer{updateTicketTypePrice: h}
}

func (c *TicketTypePriceChangedConsumer) Handle(ctx context.Context, e eventsintegrationevents.TicketTypePriceChangedIntegrationEvent) error {
	return c.updateTicketTypePrice.Handle(ctx, updatetickettypeprice.Command{
		TicketTypeID: e.TicketTypeID,
		Price:        e.Price,
	})
}
