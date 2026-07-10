package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	getuserperms "github.com/llannillo/mm/modules/users/internal/app/queries/get_user_permissions"
)

const selectUserPermissionsQuery = `
	SELECT u.id, rp.permission_code
	FROM users.users u
	JOIN users.user_roles ur ON ur.user_id = u.id
	JOIN users.role_permissions rp ON rp.role_name = ur.role_name
	WHERE u.identity_id = $1
`

type PermissionsReader struct {
	db *pgxpool.Pool
}

func NewPermissionsReader(db *pgxpool.Pool) *PermissionsReader {
	return &PermissionsReader{db: db}
}

func (r *PermissionsReader) GetUserPermissions(ctx context.Context, identityID string) (getuserperms.Result, error) {
	rows, err := r.db.Query(ctx, selectUserPermissionsQuery, identityID)
	if err != nil {
		return getuserperms.Result{}, fmt.Errorf("get user permissions: %w", err)
	}
	defer rows.Close()

	var result getuserperms.Result
	for rows.Next() {
		var permCode string
		if err := rows.Scan(&result.UserID, &permCode); err != nil {
			return getuserperms.Result{}, fmt.Errorf("scan user permissions: %w", err)
		}
		result.Permissions = append(result.Permissions, permCode)
	}
	return result, rows.Err()
}
