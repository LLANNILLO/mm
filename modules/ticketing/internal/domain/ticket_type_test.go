package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestTicketType(t *testing.T, quantity int64) *TicketType {
	t.Helper()

	return NewTicketType(uuid.New(), uuid.New(), "General Admission", 1000, "USD", quantity)
}

func TestTicketType_UpdateQuantity_ReturnsError_WhenExceedsAvailable(t *testing.T) {
	tt := newTestTicketType(t, 10)

	err := tt.UpdateQuantity(11)

	assert.ErrorIs(t, err, ErrTicketTypeInsufficientQuantity)
}

func TestTicketType_UpdateQuantity_DecrementsAvailableQuantity(t *testing.T) {
	tt := newTestTicketType(t, 10)

	err := tt.UpdateQuantity(4)

	require.NoError(t, err)
	assert.Equal(t, int64(6), tt.AvailableQuantity())
}

func TestTicketType_UpdateQuantity_RaisesDomainEvent_WhenSoldOut(t *testing.T) {
	tt := newTestTicketType(t, 5)

	err := tt.UpdateQuantity(5)

	require.NoError(t, err)
	domainEvent := assertDomainEventPublished[TicketTypeSoldOutDomainEvent](t, tt)
	assert.Equal(t, tt.ID(), domainEvent.TicketTypeID)
}

func TestTicketType_UpdateQuantity_DoesNotRaiseDomainEvent_WhenNotSoldOut(t *testing.T) {
	tt := newTestTicketType(t, 10)

	err := tt.UpdateQuantity(4)

	require.NoError(t, err)
	assertNoDomainEventPublished[TicketTypeSoldOutDomainEvent](t, tt)
}
