package publishevent

import "github.com/google/uuid"

type Command struct {
	EventID uuid.UUID
}
