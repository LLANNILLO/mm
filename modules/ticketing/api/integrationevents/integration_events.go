// Package integrationevents holds the Ticketing module's public,
// asynchronous contract. Other modules may depend on these types (via
// eventbus.Subscribe) to react to what happened in Ticketing, but must
// never depend on modules/ticketing's internal packages.
package integrationevents

import "github.com/google/uuid"

// EventPaymentsRefundedIntegrationEvent is published once every payment for
// a cancelled event has been refunded.
type EventPaymentsRefundedIntegrationEvent struct {
	EventID uuid.UUID
}

func (EventPaymentsRefundedIntegrationEvent) IsIntegrationEvent() {}

// EventTicketsArchivedIntegrationEvent is published once every ticket for a
// cancelled event has been archived.
type EventTicketsArchivedIntegrationEvent struct {
	EventID uuid.UUID
}

func (EventTicketsArchivedIntegrationEvent) IsIntegrationEvent() {}
