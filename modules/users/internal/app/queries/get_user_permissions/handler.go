package getuserperms

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const query = `
	SELECT u.id, rp.permission_code
	FROM users.users u
	JOIN users.user_roles ur ON ur.user_id = u.id
	JOIN users.role_permissions rp ON rp.role_name = ur.role_name
	WHERE u.identity_id = $1
`

type Result struct {
	UserID      uuid.UUID
	Permissions []string
}

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Handle(ctx context.Context, identityID string) (Result, error) {
	rows, err := h.db.Query(ctx, query, identityID)
	if err != nil {
		return Result{}, fmt.Errorf("get user permissions: %w", err)
	}
	defer rows.Close()

	var result Result
	for rows.Next() {
		var permCode string
		if err := rows.Scan(&result.UserID, &permCode); err != nil {
			return Result{}, fmt.Errorf("scan user permissions: %w", err)
		}
		result.Permissions = append(result.Permissions, permCode)
	}
	return result, rows.Err()
}
