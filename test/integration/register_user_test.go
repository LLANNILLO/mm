package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// registerUser posts to /users/register and returns the decoded response.
func registerUser(t *testing.T, email, password string) (id uuid.UUID, statusCode int) {
	t.Helper()

	body, err := json.Marshal(map[string]string{
		"email":      email,
		"password":   password,
		"first_name": "Jane",
		"last_name":  "Doe",
	})
	require.NoError(t, err)

	resp, err := httpClient.Post(baseURL+"/users/register", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return uuid.Nil, resp.StatusCode
	}

	var decoded struct {
		ID uuid.UUID `json:"id"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&decoded))
	return decoded.ID, resp.StatusCode
}

func TestRegisterUser_ReturnsAccessToken_WhenUserIsRegistered(t *testing.T) {
	email := fmt.Sprintf("%s@test.com", uuid.NewString())
	password := "Sup3rSecret!"

	id, status := registerUser(t, email, password)

	require.Equal(t, http.StatusCreated, status)
	require.NotEqual(t, uuid.Nil, id)

	accessToken, err := getAccessToken(email, password)

	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)
}
