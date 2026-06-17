package domain

import "github.com/google/uuid"

type UserRegisteredDomainEvent struct {
	UserID uuid.UUID
}

type UserProfileUpdatedDomainEvent struct {
	UserID    uuid.UUID
	FirstName string
	LastName  string
}

func (UserRegisteredDomainEvent) IsDomainEvent()     {}
func (UserProfileUpdatedDomainEvent) IsDomainEvent() {}
