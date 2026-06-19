package usersapi

import "github.com/google/uuid"

// UserRegisteredIntegrationEvent is published when a new user registers.
// It is the public cross-module contract for the users module.
type UserRegisteredIntegrationEvent struct {
	UserID    uuid.UUID
	Email     string
	FirstName string
	LastName  string
}

func (UserRegisteredIntegrationEvent) IsIntegrationEvent() {}

// UserProfileUpdatedIntegrationEvent is published when a user updates their profile.
type UserProfileUpdatedIntegrationEvent struct {
	UserID    uuid.UUID
	FirstName string
	LastName  string
}

func (UserProfileUpdatedIntegrationEvent) IsIntegrationEvent() {}
