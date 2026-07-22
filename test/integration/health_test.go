package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealth_ReturnsOK_WhenInfrastructureIsUp(t *testing.T) {
	resp, err := httpClient.Get(baseURL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body struct {
		Status string `json:"status"`
		Checks map[string]struct {
			Status string `json:"status"`
		} `json:"checks"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	assert.Equal(t, "Healthy", body.Status)
	assert.Equal(t, "Healthy", body.Checks["postgres"].Status)
	assert.Equal(t, "Healthy", body.Checks["valkey"].Status)
}
