package domain

import "github.com/google/uuid"

type User struct {
	entity
	ID         uuid.UUID
	Email      string
	FirstName  string
	LastName   string
	IdentityID string
}

func NewUser(email, firstName, lastName, identityID string) *User {
	u := &User{
		ID:         uuid.New(),
		Email:      email,
		FirstName:  firstName,
		LastName:   lastName,
		IdentityID: identityID,
	}
	u.raise(UserRegisteredDomainEvent{UserID: u.ID})
	return u
}

func (u *User) UpdateProfile(firstName, lastName string) {
	if u.FirstName == firstName && u.LastName == lastName {
		return
	}
	u.FirstName = firstName
	u.LastName = lastName
	u.raise(UserProfileUpdatedDomainEvent{UserID: u.ID, FirstName: firstName, LastName: lastName})
}
