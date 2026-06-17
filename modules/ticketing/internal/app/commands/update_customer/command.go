package updatecustomer

import "github.com/google/uuid"

type Command struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
}
