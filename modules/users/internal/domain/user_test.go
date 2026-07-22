package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestUser(t *testing.T) *User {
	t.Helper()

	return NewUser("jane@example.com", "Jane", "Doe", "identity-1")
}

func TestNewUser_RaisesDomainEvent(t *testing.T) {
	u := newTestUser(t)

	domainEvent := assertDomainEventPublished[UserRegisteredDomainEvent](t, u)
	assert.Equal(t, u.ID(), domainEvent.UserID)
}

func TestUser_UpdateProfile_RaisesDomainEvent_WhenChanged(t *testing.T) {
	u := newTestUser(t)
	u.ClearDomainEvents()

	u.UpdateProfile("Janet", "Smith")

	assert.Equal(t, "Janet", u.FirstName())
	assert.Equal(t, "Smith", u.LastName())
	domainEvent := assertDomainEventPublished[UserProfileUpdatedDomainEvent](t, u)
	assert.Equal(t, u.ID(), domainEvent.UserID)
	assert.Equal(t, "Janet", domainEvent.FirstName)
	assert.Equal(t, "Smith", domainEvent.LastName)
}

func TestUser_UpdateProfile_DoesNotRaiseDomainEvent_WhenUnchanged(t *testing.T) {
	u := newTestUser(t)
	u.ClearDomainEvents()

	u.UpdateProfile(u.FirstName(), u.LastName())

	assertNoDomainEventPublished[UserProfileUpdatedDomainEvent](t, u)
}
