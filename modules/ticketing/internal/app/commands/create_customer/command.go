package createcustomer

import "github.com/google/uuid"

type Command struct {
	ID        uuid.UUID
	Email     string
	FirstName string
	LastName  string
}
