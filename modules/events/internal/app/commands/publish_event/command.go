package publishevent

import (
	"github.com/google/uuid"
	"github.com/llannillo/mm/internal/shared/validation"
)

type Command struct {
	EventID uuid.UUID
}

func (c Command) Validate() error {
	return validation.New().
		Custom("event_id", c.EventID == uuid.Nil, "event_id is required").
		Err()
}
