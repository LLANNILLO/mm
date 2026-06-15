package events

import (
	"time"

	"github.com/google/uuid"
)

// EventPublishedIntegrationEvent is published when an event transitions to published status.
// Other modules subscribe to this to react to published events.
type EventPublishedIntegrationEvent struct {
	EventID    uuid.UUID
	OccurredAt time.Time
}

// EventCancelledIntegrationEvent is published when an event is cancelled.
type EventCancelledIntegrationEvent struct {
	EventID    uuid.UUID
	OccurredAt time.Time
}
