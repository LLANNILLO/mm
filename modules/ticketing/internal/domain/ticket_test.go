package domain

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestTicket(t *testing.T) (*Ticket, uuid.UUID) {
	t.Helper()

	customerID := uuid.New()
	order := NewOrder(customerID)
	ticketType := newTestTicketType(t, 10)

	return NewTicket(order, ticketType), customerID
}

func TestNewTicket_RaisesDomainEvent(t *testing.T) {
	ticket, _ := newTestTicket(t)

	domainEvent := assertDomainEventPublished[TicketCreatedDomainEvent](t, ticket)
	assert.Equal(t, ticket.ID(), domainEvent.TicketID)
	assert.Equal(t, ticket.EventID(), domainEvent.EventID)
	assert.True(t, strings.HasPrefix(ticket.Code(), "tc_"))
}

func TestTicket_Archive_RaisesDomainEvent_WhenNotArchived(t *testing.T) {
	ticket, _ := newTestTicket(t)

	ticket.Archive()

	assert.True(t, ticket.Archived())
	domainEvent := assertDomainEventPublished[TicketArchivedDomainEvent](t, ticket)
	assert.Equal(t, ticket.ID(), domainEvent.TicketID)
	assert.Equal(t, ticket.Code(), domainEvent.Code)
}

func TestTicket_Archive_DoesNotRaiseDomainEvent_WhenAlreadyArchived(t *testing.T) {
	ticket, _ := newTestTicket(t)
	ticket.Archive()
	ticket.ClearDomainEvents()

	ticket.Archive()

	assertNoDomainEventPublished[TicketArchivedDomainEvent](t, ticket)
}

func TestTicket_CheckIn_RaisesDomainEvent_WhenValid(t *testing.T) {
	ticket, customerID := newTestTicket(t)

	err := ticket.CheckIn(customerID)

	require.NoError(t, err)
	require.NotNil(t, ticket.UsedAtUtc())
	domainEvent := assertDomainEventPublished[TicketCheckedInDomainEvent](t, ticket)
	assert.Equal(t, ticket.ID(), domainEvent.TicketID)
	assert.Equal(t, ticket.EventID(), domainEvent.EventID)
}

func TestTicket_CheckIn_RaisesDomainEvent_WhenWrongCustomer(t *testing.T) {
	ticket, _ := newTestTicket(t)

	err := ticket.CheckIn(uuid.New())

	assert.ErrorIs(t, err, ErrTicketCheckInInvalid)
	domainEvent := assertDomainEventPublished[TicketCheckInInvalidDomainEvent](t, ticket)
	assert.Equal(t, ticket.ID(), domainEvent.TicketID)
}

func TestTicket_CheckIn_RaisesDomainEvent_WhenAlreadyCheckedIn(t *testing.T) {
	ticket, customerID := newTestTicket(t)
	require.NoError(t, ticket.CheckIn(customerID))
	ticket.ClearDomainEvents()

	err := ticket.CheckIn(customerID)

	assert.ErrorIs(t, err, ErrTicketAlreadyCheckedIn)
	domainEvent := assertDomainEventPublished[TicketCheckInDuplicateDomainEvent](t, ticket)
	assert.Equal(t, ticket.ID(), domainEvent.TicketID)
}
