// Package integrationevents holds the Events module's public, asynchronous
// contract. Other modules may depend on these types (via eventbus.Subscribe)
// to react to what happened in Events, but must never depend on
// modules/events/api's synchronous EventsAPI interface.
package integrationevents

import (
	"time"

	"github.com/google/uuid"
)

// EventCreatedIntegrationEvent is published when a new event is created.
// It is the public cross-module contract for the events module.
type EventCreatedIntegrationEvent struct {
	EventID     uuid.UUID
	CategoryID  uuid.UUID
	Title       string
	Description *string
	Location    *string
	StartsAtUtc time.Time
	EndsAtUtc   *time.Time
}

func (EventCreatedIntegrationEvent) IsIntegrationEvent() {}

// EventRescheduledIntegrationEvent is published when an event's schedule changes.
type EventRescheduledIntegrationEvent struct {
	EventID     uuid.UUID
	StartsAtUtc time.Time
	EndsAtUtc   *time.Time
}

func (EventRescheduledIntegrationEvent) IsIntegrationEvent() {}

// TicketTypeCreatedIntegrationEvent is published when a new ticket type is created.
type TicketTypeCreatedIntegrationEvent struct {
	TicketTypeID uuid.UUID
	EventID      uuid.UUID
	Name         string
	Price        int64
	Currency     string
	Quantity     int64
}

func (TicketTypeCreatedIntegrationEvent) IsIntegrationEvent() {}

// TicketTypePriceChangedIntegrationEvent is published when a ticket type's price changes.
type TicketTypePriceChangedIntegrationEvent struct {
	TicketTypeID uuid.UUID
	Price        int64
}

func (TicketTypePriceChangedIntegrationEvent) IsIntegrationEvent() {}

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
