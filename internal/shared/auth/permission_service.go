package auth

import (
	"context"

	"github.com/google/uuid"
)

type PermissionService interface {
	GetUserPermissions(ctx context.Context, identityID string) (userID uuid.UUID, permissions []string, err error)
}
