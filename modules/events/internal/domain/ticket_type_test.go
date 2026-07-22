package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestTicketType(t *testing.T) *TicketType {
	t.Helper()

	tt, err := NewTicketType(uuid.New(), "General Admission", 1000, "USD", 100)
	require.NoError(t, err)

	return tt
}

func TestNewTicketType_ReturnsError_WhenNameEmpty(t *testing.T) {
	_, err := NewTicketType(uuid.New(), "", 1000, "USD", 100)

	assert.ErrorIs(t, err, ErrTicketTypeNameEmpty)
}

func TestNewTicketType_ReturnsError_WhenPriceNegative(t *testing.T) {
	_, err := NewTicketType(uuid.New(), "General Admission", -1, "USD", 100)

	assert.ErrorIs(t, err, ErrTicketTypeInvalidPrice)
}

func TestNewTicketType_RaisesDomainEvent_WhenCreated(t *testing.T) {
	tt := newTestTicketType(t)

	domainEvent := assertDomainEventPublished[TicketTypeCreatedDomainEvent](t, tt)
	assert.Equal(t, tt.ID(), domainEvent.TicketTypeID)
}

func TestTicketType_UpdatePrice_ReturnsError_WhenPriceNegative(t *testing.T) {
	tt := newTestTicketType(t)

	err := tt.UpdatePrice(-1)

	assert.ErrorIs(t, err, ErrTicketTypeInvalidPrice)
}

func TestTicketType_UpdatePrice_RaisesDomainEvent_WhenPriceChanged(t *testing.T) {
	tt := newTestTicketType(t)
	tt.ClearDomainEvents()

	err := tt.UpdatePrice(2000)

	require.NoError(t, err)
	domainEvent := assertDomainEventPublished[TicketTypePriceChangedDomainEvent](t, tt)
	assert.Equal(t, tt.ID(), domainEvent.TicketTypeID)
	assert.Equal(t, int64(2000), domainEvent.Price)
}

func TestTicketType_UpdatePrice_DoesNotRaiseDomainEvent_WhenUnchanged(t *testing.T) {
	tt := newTestTicketType(t)
	tt.ClearDomainEvents()

	err := tt.UpdatePrice(tt.Price())

	require.NoError(t, err)
	assert.Empty(t, tt.DomainEvents())
}
