package eventhandlers

import (
	"context"

	"github.com/llannillo/mm/internal/shared/eventbus"
	"github.com/llannillo/mm/modules/ticketing/api/integrationevents"
	"github.com/llannillo/mm/modules/ticketing/internal/domain"
)

// PaymentsRefundedIntegrationEventPublisher and TicketsArchivedIntegrationEventPublisher
// republish this module's local completion domain events as its public
// integration events, so the cancel-event saga (in modules/events) can
// observe them without depending on this module's internals.

type PaymentsRefundedIntegrationEventPublisher struct {
	eventBus eventbus.EventBus
}

func NewPaymentsRefundedIntegrationEventPublisher(eventBus eventbus.EventBus) *PaymentsRefundedIntegrationEventPublisher {
	return &PaymentsRefundedIntegrationEventPublisher{eventBus: eventBus}
}

func (h *PaymentsRefundedIntegrationEventPublisher) Handle(ctx context.Context, e domain.EventPaymentsRefundedDomainEvent) error {
	return h.eventBus.Publish(ctx, integrationevents.EventPaymentsRefundedIntegrationEvent{EventID: e.EventID})
}

type TicketsArchivedIntegrationEventPublisher struct {
	eventBus eventbus.EventBus
}

func NewTicketsArchivedIntegrationEventPublisher(eventBus eventbus.EventBus) *TicketsArchivedIntegrationEventPublisher {
	return &TicketsArchivedIntegrationEventPublisher{eventBus: eventBus}
}

func (h *TicketsArchivedIntegrationEventPublisher) Handle(ctx context.Context, e domain.EventTicketsArchivedDomainEvent) error {
	return h.eventBus.Publish(ctx, integrationevents.EventTicketsArchivedIntegrationEvent{EventID: e.EventID})
}
