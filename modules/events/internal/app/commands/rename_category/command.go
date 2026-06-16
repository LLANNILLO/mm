package renamecategory

import (
	"github.com/google/uuid"
	"github.com/llannillo/mm/internal/shared/validation"
)

type Command struct {
	CategoryID uuid.UUID
	Name       string
}

func (c Command) Validate() error {
	return validation.New().
		Custom("category_id", c.CategoryID == uuid.Nil, "category_id is required").
		Required("name", c.Name).
		Err()
}
