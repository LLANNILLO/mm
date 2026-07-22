package outbound

import (
	"context"

	"github.com/google/uuid"
)

// Step identifies one of the parallel branches the cancel-event saga waits
// on before declaring a cancellation complete.
type Step uint8

const (
	StepPaymentsRefunded Step = 1 << iota
	StepTicketsArchived
)

// AllSteps is the value MarkStepComplete's result is compared against to
// know every branch has reported in.
const AllSteps = StepPaymentsRefunded | StepTicketsArchived

// CancelEventSagaRepository persists the cancel-event saga's correlation
// state: one row per in-flight cancellation, tracking which of Ticketing's
// two parallel completion steps (refund, archive) have arrived so far.
type CancelEventSagaRepository interface {
	Start(ctx context.Context, eventID uuid.UUID) error
	MarkStepComplete(ctx context.Context, eventID uuid.UUID, step Step) (completedSteps Step, err error)
	Delete(ctx context.Context, eventID uuid.UUID) error
}
