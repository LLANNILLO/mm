package createevent

import "time"

type Command struct {
	Title       string
	Description *string
	Location    *string
	StartsAtUtc time.Time
	EndsAtUtc   *time.Time
}
