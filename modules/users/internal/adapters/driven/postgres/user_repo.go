package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/llannillo/mm/internal/shared/events"
	store "github.com/llannillo/mm/modules/users/internal/adapters/driven/postgres/generated"
	"github.com/llannillo/mm/modules/users/internal/domain"
)

type UserRepository struct {
	queries    *store.Queries
	dispatcher *events.Dispatcher
}

func NewUserRepository(q *store.Queries, d *events.Dispatcher) *UserRepository {
	return &UserRepository{queries: q, dispatcher: d}
}

func (r *UserRepository) Insert(ctx context.Context, u *domain.User) error {
	_, err := r.queries.InsertUser(ctx, store.InsertUserParams{
		ID:         u.ID(),
		Email:      u.Email(),
		FirstName:  u.FirstName(),
		LastName:   u.LastName(),
		IdentityID: u.IdentityID(),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrEmailAlreadyTaken
		}
		return fmt.Errorf("insert user: %w", err)
	}

	if err := r.queries.InsertUserRole(ctx, store.InsertUserRoleParams{
		UserID:   u.ID(),
		RoleName: domain.RoleMember,
	}); err != nil {
		return fmt.Errorf("insert user role: %w", err)
	}

	domainEvents := u.DomainEvents()
	u.ClearDomainEvents()
	return r.dispatcher.Dispatch(ctx, domainEvents)
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row, err := r.queries.SelectUserForUpdate(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return rehydrateUser(row), nil
}

func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	for _, e := range u.DomainEvents() {
		var err error
		switch e.(type) {
		case domain.UserProfileUpdatedDomainEvent:
			err = r.queries.UpdateUserProfile(ctx, store.UpdateUserProfileParams{
				ID:        u.ID(),
				FirstName: u.FirstName(),
				LastName:  u.LastName(),
			})
		}
		if err != nil {
			return err
		}
	}
	domainEvents := u.DomainEvents()
	u.ClearDomainEvents()
	return r.dispatcher.Dispatch(ctx, domainEvents)
}

func rehydrateUser(row store.UsersUser) *domain.User {
	return domain.RehydrateUser(row.ID, row.Email, row.FirstName, row.LastName, row.IdentityID)
}
