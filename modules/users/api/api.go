package usersapi

import (
	"context"

	"github.com/google/uuid"
)

type UsersAPI interface {
	GetUser(ctx context.Context, id uuid.UUID) (*UserResponse, error)
}

type UserResponse struct {
	ID        uuid.UUID
	Email     string
	FirstName string
	LastName  string
}
