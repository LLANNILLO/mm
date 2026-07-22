package eventcancellation

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/llannillo/mm/internal/shared/eventbus"
	"github.com/llannillo/mm/modules/events/api/integrationevents"
	"github.com/llannillo/mm/modules/events/internal/ports/outbound"
	ticketingintegrationevents "github.com/llannillo/mm/modules/ticketing/api/integrationevents"
)

// PaymentsRefundedHandler and TicketsArchivedHandler each mark their branch
// of the cancel-event saga complete and, once both have reported in,
// publish the saga's completion event and drop its state.

type PaymentsRefundedHandler struct {
	sagaRepo outbound.CancelEventSagaRepository
	eventBus eventbus.EventBus
}

func NewPaymentsRefundedHandler(sagaRepo outbound.CancelEventSagaRepository, eventBus eventbus.EventBus) *PaymentsRefundedHandler {
	return &PaymentsRefundedHandler{sagaRepo: sagaRepo, eventBus: eventBus}
}

func (h *PaymentsRefundedHandler) Handle(ctx context.Context, e ticketingintegrationevents.EventPaymentsRefundedIntegrationEvent) error {
	return markStepComplete(ctx, h.sagaRepo, h.eventBus, e.EventID, outbound.StepPaymentsRefunded)
}

type TicketsArchivedHandler struct {
	sagaRepo outbound.CancelEventSagaRepository
	eventBus eventbus.EventBus
}

func NewTicketsArchivedHandler(sagaRepo outbound.CancelEventSagaRepository, eventBus eventbus.EventBus) *TicketsArchivedHandler {
	return &TicketsArchivedHandler{sagaRepo: sagaRepo, eventBus: eventBus}
}

func (h *TicketsArchivedHandler) Handle(ctx context.Context, e ticketingintegrationevents.EventTicketsArchivedIntegrationEvent) error {
	return markStepComplete(ctx, h.sagaRepo, h.eventBus, e.EventID, outbound.StepTicketsArchived)
}

func markStepComplete(
	ctx context.Context,
	sagaRepo outbound.CancelEventSagaRepository,
	eventBus eventbus.EventBus,
	eventID uuid.UUID,
	step outbound.Step,
) error {
	completed, err := sagaRepo.MarkStepComplete(ctx, eventID, step)
	if err != nil {
		return fmt.Errorf("mark cancel event saga step complete: %w", err)
	}
	if completed != outbound.AllSteps {
		return nil
	}
	if err := eventBus.Publish(ctx, integrationevents.EventCancellationCompletedIntegrationEvent{EventID: eventID}); err != nil {
		return fmt.Errorf("publish event cancellation completed: %w", err)
	}
	return sagaRepo.Delete(ctx, eventID)
}
