package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	store "github.com/llannillo/mm/modules/users/internal/adapters/driven/postgres/generated"
	getuser "github.com/llannillo/mm/modules/users/internal/app/queries/get_user"
	"github.com/llannillo/mm/modules/users/internal/domain"
)

type UserReader struct {
	queries *store.Queries
}

func NewUserReader(q *store.Queries) *UserReader {
	return &UserReader{queries: q}
}

func (r *UserReader) GetUser(ctx context.Context, id uuid.UUID) (*getuser.Response, error) {
	row, err := r.queries.SelectUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &getuser.Response{
		ID:        row.ID,
		Email:     row.Email,
		FirstName: row.FirstName,
		LastName:  row.LastName,
	}, nil
}
