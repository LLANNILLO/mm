package createtickettype

import "github.com/google/uuid"

type Command struct {
	ID       uuid.UUID
	EventID  uuid.UUID
	Name     string
	Price    int64
	Currency string
	Quantity int64
}
