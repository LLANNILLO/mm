package eventcancellation

import (
	"context"
	"fmt"

	"github.com/llannillo/mm/internal/shared/eventbus"
	"github.com/llannillo/mm/modules/events/api/integrationevents"
	"github.com/llannillo/mm/modules/events/internal/ports/outbound"
)

// StartedHandler begins tracking a cancellation once the event has been
// cancelled, then tells other modules to start their own reaction work.
type StartedHandler struct {
	sagaRepo outbound.CancelEventSagaRepository
	eventBus eventbus.EventBus
}

func NewStartedHandler(sagaRepo outbound.CancelEventSagaRepository, eventBus eventbus.EventBus) *StartedHandler {
	return &StartedHandler{sagaRepo: sagaRepo, eventBus: eventBus}
}

func (h *StartedHandler) Handle(ctx context.Context, e integrationevents.EventCanceledIntegrationEvent) error {
	if err := h.sagaRepo.Start(ctx, e.EventID); err != nil {
		return fmt.Errorf("start cancel event saga: %w", err)
	}
	return h.eventBus.Publish(ctx, integrationevents.EventCancellationStartedIntegrationEvent{EventID: e.EventID})
}
