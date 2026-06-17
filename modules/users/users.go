package users

import "github.com/google/uuid"

// UserRegisteredIntegrationEvent is published when a user registers.
// Other modules subscribe to this for async cross-module communication.
type UserRegisteredIntegrationEvent struct {
	UserID uuid.UUID
}
