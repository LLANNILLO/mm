package getuserperms

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Result struct {
	UserID      uuid.UUID
	Permissions []string
}

type PermissionsReader interface {
	GetUserPermissions(ctx context.Context, identityID string) (Result, error)
}

type Handler struct {
	reader PermissionsReader
}

func NewHandler(reader PermissionsReader) *Handler {
	return &Handler{reader: reader}
}

func (h *Handler) Handle(ctx context.Context, q Query) (Result, error) {
	result, err := h.reader.GetUserPermissions(ctx, q.IdentityID)
	if err != nil {
		return Result{}, fmt.Errorf("get user permissions: %w", err)
	}
	return result, nil
}
