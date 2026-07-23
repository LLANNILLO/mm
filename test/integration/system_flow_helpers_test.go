package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// authedRequest builds a JSON request carrying a Bearer token — every
// endpoint except POST /users/register and GET /health requires one.
func authedRequest(t *testing.T, method, path, token string, body any) *http.Response {
	t.Helper()

	raw, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(method, baseURL+path, bytes.NewReader(raw))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	return resp
}

func createCategory(t *testing.T, token, name string) uuid.UUID {
	t.Helper()

	resp := authedRequest(t, http.MethodPost, "/categories", token, map[string]string{"name": name})
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var decoded struct {
		ID uuid.UUID `json:"id"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&decoded))
	return decoded.ID
}

func createEvent(t *testing.T, token string, categoryID uuid.UUID, title string) uuid.UUID {
	t.Helper()

	resp := authedRequest(t, http.MethodPost, "/events", token, map[string]any{
		"category_id":   categoryID,
		"title":         title,
		"starts_at_utc": time.Now().Add(24 * time.Hour),
	})
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var decoded struct {
		ID uuid.UUID `json:"id"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&decoded))
	return decoded.ID
}

func createTicketType(t *testing.T, token string, eventID uuid.UUID, name string, quantity int64) uuid.UUID {
	t.Helper()

	resp := authedRequest(t, http.MethodPost, "/ticket-types", token, map[string]any{
		"event_id": eventID,
		"name":     name,
		"price":    1000,
		"currency": "USD",
		"quantity": quantity,
	})
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var decoded struct {
		ID uuid.UUID `json:"id"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&decoded))
	return decoded.ID
}

// tryAddItemToCart returns true once the request succeeds. Used to poll for
// the eventually-consistent moment a customer and/or ticket type replica has
// propagated into the ticketing module.
func tryAddItemToCart(t *testing.T, token string, customerID, ticketTypeID uuid.UUID, quantity int64) bool {
	t.Helper()

	resp := authedRequest(t, http.MethodPut, "/carts/add", token, map[string]any{
		"customer_id":    customerID,
		"ticket_type_id": ticketTypeID,
		"quantity":       quantity,
	})
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// registerTestUser registers a new user and returns its ID plus a valid
// access token for it, obtained the same way BaseIntegrationTest.
// GetAccessTokenAsync does in the C# reference.
func registerTestUser(t *testing.T) (id uuid.UUID, token string) {
	t.Helper()

	email := fmt.Sprintf("%s@test.com", uuid.NewString())
	password := "Sup3rSecret!"

	id, status := registerUser(t, email, password)
	require.Equal(t, http.StatusCreated, status)

	token, err := getAccessToken(email, password)
	require.NoError(t, err)

	return id, token
}
