package domain

import "github.com/google/uuid"

type User struct {
	entity
	id         uuid.UUID
	email      string
	firstName  string
	lastName   string
	identityID string
}

func (u *User) ID() uuid.UUID      { return u.id }
func (u *User) Email() string      { return u.email }
func (u *User) FirstName() string  { return u.firstName }
func (u *User) LastName() string   { return u.lastName }
func (u *User) IdentityID() string { return u.identityID }

func NewUser(email, firstName, lastName, identityID string) *User {
	u := &User{
		id:         uuid.New(),
		email:      email,
		firstName:  firstName,
		lastName:   lastName,
		identityID: identityID,
	}
	u.raise(UserRegisteredDomainEvent{UserID: u.id})
	return u
}

// RehydrateUser reconstructs a User from persisted state without raising
// domain events. Only the UserRepository may call this.
func RehydrateUser(id uuid.UUID, email, firstName, lastName, identityID string) *User {
	return &User{id: id, email: email, firstName: firstName, lastName: lastName, identityID: identityID}
}

func (u *User) UpdateProfile(firstName, lastName string) {
	if u.firstName == firstName && u.lastName == lastName {
		return
	}
	u.firstName = firstName
	u.lastName = lastName
	u.raise(UserProfileUpdatedDomainEvent{UserID: u.id, FirstName: firstName, LastName: lastName})
}
