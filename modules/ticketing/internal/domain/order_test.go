package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOrder_RaisesDomainEvent(t *testing.T) {
	o := NewOrder(uuid.New())

	domainEvent := assertDomainEventPublished[OrderCreatedDomainEvent](t, o)
	assert.Equal(t, o.ID(), domainEvent.OrderID)
}

func TestOrder_AddItem_AccumulatesTotalPriceAndCurrency(t *testing.T) {
	o := NewOrder(uuid.New())
	ticketType := newTestTicketType(t, 10)

	err := o.AddItem(ticketType, 3)

	require.NoError(t, err)
	require.Len(t, o.Items(), 1)
	assert.Equal(t, int64(3000), o.TotalPrice())
	assert.Equal(t, "USD", o.Currency())
	assert.Equal(t, ticketType.ID(), o.Items()[0].TicketTypeID())
	assert.Equal(t, int64(3), o.Items()[0].Quantity())
}

func TestOrder_IssueTickets_RaisesDomainEvent(t *testing.T) {
	o := NewOrder(uuid.New())

	err := o.IssueTickets()

	require.NoError(t, err)
	assert.True(t, o.TicketsIssued())
	domainEvent := assertDomainEventPublished[OrderTicketsIssuedDomainEvent](t, o)
	assert.Equal(t, o.ID(), domainEvent.OrderID)
}

func TestOrder_IssueTickets_ReturnsError_WhenAlreadyIssued(t *testing.T) {
	o := NewOrder(uuid.New())
	require.NoError(t, o.IssueTickets())

	err := o.IssueTickets()

	assert.ErrorIs(t, err, ErrOrderTicketsAlreadyIssued)
}
