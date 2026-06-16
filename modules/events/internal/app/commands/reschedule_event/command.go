package rescheduleevent

import (
	"time"

	"github.com/google/uuid"
	"github.com/llannillo/mm/internal/shared/validation"
)

type Command struct {
	EventID     uuid.UUID
	StartsAtUtc time.Time
	EndsAtUtc   *time.Time
}

func (c Command) Validate() error {
	return validation.New().
		Custom("event_id", c.EventID == uuid.Nil, "event_id is required").
		Custom("starts_at_utc", c.StartsAtUtc.IsZero(), "starts_at_utc is required").
		Custom("ends_at_utc", c.EndsAtUtc != nil && !c.EndsAtUtc.After(c.StartsAtUtc), "ends_at_utc must be after starts_at_utc").
		Err()
}
