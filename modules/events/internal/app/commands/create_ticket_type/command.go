package createtickettype

import "github.com/google/uuid"

type Command struct {
	EventID  uuid.UUID
	Name     string
	Price    int64
	Currency string
	Quantity int64
}
