package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestSystemFlow_UserCanAddItemToCartAfterRegistrationAndEventCreation is the
// Go equivalent of the C# reference's RegisterUserTests + AddItemToCartTests
// combined into one flow. Both propagation paths built this session —
// Users -> Ticketing (Customer) and Events -> Ticketing (Event/TicketType) —
// only become observable through the same endpoint, PUT /carts/add: there's
// no dedicated "GET customer" or "GET replica ticket type" route to poll
// individually, so proving the cart call eventually succeeds proves both
// outbox/inbox pipelines landed.
func TestSystemFlow_UserCanAddItemToCartAfterRegistrationAndEventCreation(t *testing.T) {
	customerID, token := registerTestUser(t)

	categoryID := createCategory(t, token, "Music-"+uuid.NewString())
	eventID := createEvent(t, token, categoryID, "Concert-"+uuid.NewString())
	ticketTypeID := createTicketType(t, token, eventID, "General Admission", 100)

	poll(t, 15*time.Second, func() bool {
		return tryAddItemToCart(t, token, customerID, ticketTypeID, 2)
	})
}
