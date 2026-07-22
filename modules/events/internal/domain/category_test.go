package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCategory(t *testing.T) *Category {
	t.Helper()

	c, err := NewCategory("Music")
	require.NoError(t, err)

	return c
}

func TestNewCategory_ReturnsError_WhenNameEmpty(t *testing.T) {
	_, err := NewCategory("")

	assert.ErrorIs(t, err, ErrCategoryNameEmpty)
}

func TestNewCategory_RaisesDomainEvent_WhenCreated(t *testing.T) {
	c := newTestCategory(t)

	domainEvent := assertDomainEventPublished[CategoryCreatedDomainEvent](t, c)
	assert.Equal(t, c.ID(), domainEvent.CategoryID)
}

func TestCategory_Archive_RaisesDomainEvent_WhenArchived(t *testing.T) {
	c := newTestCategory(t)

	err := c.Archive()

	require.NoError(t, err)
	domainEvent := assertDomainEventPublished[CategoryArchivedDomainEvent](t, c)
	assert.Equal(t, c.ID(), domainEvent.CategoryID)
}

func TestCategory_Archive_ReturnsError_WhenAlreadyArchived(t *testing.T) {
	c := newTestCategory(t)
	require.NoError(t, c.Archive())

	err := c.Archive()

	assert.ErrorIs(t, err, ErrCategoryAlreadyArchived)
}

func TestCategory_ChangeName_ReturnsError_WhenNameEmpty(t *testing.T) {
	c := newTestCategory(t)

	err := c.ChangeName("")

	assert.ErrorIs(t, err, ErrCategoryNameEmpty)
}

func TestCategory_ChangeName_RaisesDomainEvent_WhenNameChanged(t *testing.T) {
	c := newTestCategory(t)
	c.ClearDomainEvents()

	err := c.ChangeName("Sports")

	require.NoError(t, err)
	domainEvent := assertDomainEventPublished[CategoryNameChangedDomainEvent](t, c)
	assert.Equal(t, c.ID(), domainEvent.CategoryID)
	assert.Equal(t, "Sports", domainEvent.Name)
}

func TestCategory_ChangeName_DoesNotRaiseDomainEvent_WhenUnchanged(t *testing.T) {
	c := newTestCategory(t)
	c.ClearDomainEvents()

	err := c.ChangeName(c.Name())

	require.NoError(t, err)
	assert.Empty(t, c.DomainEvents())
}
