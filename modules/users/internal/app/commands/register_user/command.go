package registeruser

import (
	"strings"

	"github.com/llannillo/mm/internal/shared/validation"
)

type Command struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

func (c Command) Validate() error {
	return validation.New().
		Required("email", c.Email).
		Custom("email", !strings.Contains(c.Email, "@"), "email is not valid").
		Required("password", c.Password).
		Custom("password", len(c.Password) < 6, "password must be at least 6 characters").
		Required("first_name", c.FirstName).
		Required("last_name", c.LastName).
		Err()
}
