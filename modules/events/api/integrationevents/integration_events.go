// Package integrationevents holds the Events module's public, asynchronous
// contract. Other modules may depend on these types (via eventbus.Subscribe)
// to react to what happened in Events, but must never depend on
// modules/events/api's synchronous EventsAPI interface.
package integrationevents

import "github.com/google/uuid"

// EventCanceledIntegrationEvent is published when an event is cancelled.
// It is the public cross-module contract for the events module.
type EventCanceledIntegrationEvent struct {
	EventID uuid.UUID
}

func (EventCanceledIntegrationEvent) IsIntegrationEvent() {}

// EventCancellationStartedIntegrationEvent is published by the cancel-event
// saga once it has observed EventCanceledIntegrationEvent, signalling other
// modules to begin their own cancellation-triggered work.
type EventCancellationStartedIntegrationEvent struct {
	EventID uuid.UUID
}

func (EventCancellationStartedIntegrationEvent) IsIntegrationEvent() {}

// EventCancellationCompletedIntegrationEvent is published by the cancel-event
// saga once every module has finished reacting to the cancellation.
type EventCancellationCompletedIntegrationEvent struct {
	EventID uuid.UUID
}

func (EventCancellationCompletedIntegrationEvent) IsIntegrationEvent() {}
