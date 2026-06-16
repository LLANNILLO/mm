package createcategory

import "github.com/llannillo/mm/internal/shared/validation"

type Command struct {
	Name string
}

func (c Command) Validate() error {
	return validation.New().
		Required("name", c.Name).
		Err()
}
