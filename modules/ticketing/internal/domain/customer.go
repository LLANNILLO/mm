package domain

import "github.com/google/uuid"

type Customer struct {
	id        uuid.UUID
	email     string
	firstName string
	lastName  string
}

func (c *Customer) ID() uuid.UUID     { return c.id }
func (c *Customer) Email() string     { return c.email }
func (c *Customer) FirstName() string { return c.firstName }
func (c *Customer) LastName() string  { return c.lastName }

func NewCustomer(id uuid.UUID, email, firstName, lastName string) (*Customer, error) {
	return &Customer{
		id:        id,
		email:     email,
		firstName: firstName,
		lastName:  lastName,
	}, nil
}

// RehydrateCustomer reconstructs a Customer from persisted state without
// re-running creation invariants. Only repositories may call this.
func RehydrateCustomer(id uuid.UUID, email, firstName, lastName string) *Customer {
	return &Customer{id: id, email: email, firstName: firstName, lastName: lastName}
}

func (c *Customer) Update(firstName, lastName string) {
	c.firstName = firstName
	c.lastName = lastName
}
