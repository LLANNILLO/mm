package eventhandlers

import (
	"context"

	"github.com/llannillo/mm/modules/events/internal/domain"
)

func HandleEventRescheduled(_ context.Context, _ domain.EventRescheduledDomainEvent) error {
	return nil
}
