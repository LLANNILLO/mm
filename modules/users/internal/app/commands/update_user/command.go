package updateuser

import (
	"github.com/google/uuid"
	"github.com/llannillo/mm/internal/shared/validation"
)

type Command struct {
	UserID    uuid.UUID
	FirstName string
	LastName  string
}

func (c Command) Validate() error {
	return validation.New().
		Custom("user_id", c.UserID == uuid.Nil, "user_id is required").
		Required("first_name", c.FirstName).
		Required("last_name", c.LastName).
		Err()
}
