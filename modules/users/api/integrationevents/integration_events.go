// Package integrationevents holds the Users module's public, asynchronous
// contract. Other modules may depend on these types (via eventbus.Subscribe)
// to react to what happened in Users, but must never depend on
// modules/users/api's synchronous UsersAPI interface.
package integrationevents

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
