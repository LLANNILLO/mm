package domain

import "github.com/google/uuid"

type Customer struct {
	ID        uuid.UUID
	Email     string
	FirstName string
	LastName  string
}

func NewCustomer(id uuid.UUID, email, firstName, lastName string) (*Customer, error) {
	return &Customer{
		ID:        id,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	}, nil
}

func (c *Customer) Update(firstName, lastName string) {
	c.FirstName = firstName
	c.LastName = lastName
}
