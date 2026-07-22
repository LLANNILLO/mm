package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomer_Update_ChangesFirstAndLastName(t *testing.T) {
	c, err := NewCustomer(uuid.New(), "jane@example.com", "Jane", "Doe")
	require.NoError(t, err)

	c.Update("Janet", "Smith")

	assert.Equal(t, "Janet", c.FirstName())
	assert.Equal(t, "Smith", c.LastName())
	assert.Equal(t, "jane@example.com", c.Email())
}
