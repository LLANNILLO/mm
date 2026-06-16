package createevent

import (
	"time"

	"github.com/google/uuid"
	"github.com/llannillo/mm/internal/shared/validation"
)

type Command struct {
	CategoryID  uuid.UUID
	Title       string
	Description *string
	Location    *string
	StartsAtUtc time.Time
	EndsAtUtc   *time.Time
}

func (c Command) Validate() error {
	return validation.New().
		Custom("category_id", c.CategoryID == uuid.Nil, "category_id is required").
		Required("title", c.Title).
		Custom("starts_at_utc", c.StartsAtUtc.IsZero(), "starts_at_utc is required").
		Custom("ends_at_utc", c.EndsAtUtc != nil && !c.EndsAtUtc.After(c.StartsAtUtc), "ends_at_utc must be after starts_at_utc").
		Err()
}
