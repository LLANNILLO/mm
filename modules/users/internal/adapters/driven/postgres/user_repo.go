package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/llannillo/mm/internal/shared/outbox"
	store "github.com/llannillo/mm/modules/users/internal/adapters/driven/postgres/generated"
	"github.com/llannillo/mm/modules/users/internal/domain"
)

const schema = "users"

type UserRepository struct {
	queries *store.Queries
	uow     *UnitOfWork
}

func NewUserRepository(q *store.Queries, uow *UnitOfWork) *UserRepository {
	return &UserRepository{queries: q, uow: uow}
}

func (r *UserRepository) Insert(ctx context.Context, u *domain.User) error {
	return r.uow.WithTx(ctx, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		_, err := q.InsertUser(ctx, store.InsertUserParams{
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

		if err := q.InsertUserRole(ctx, store.InsertUserRoleParams{
			UserID:   u.ID(),
			RoleName: domain.RoleMember,
		}); err != nil {
			return fmt.Errorf("insert user role: %w", err)
		}

		domainEvents := u.DomainEvents()
		u.ClearDomainEvents()
		return outbox.InsertMessages(ctx, tx, schema, domainEvents)
	})
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
	return r.uow.WithTx(ctx, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		for _, e := range u.DomainEvents() {
			var err error
			switch e.(type) {
			case domain.UserProfileUpdatedDomainEvent:
				err = q.UpdateUserProfile(ctx, store.UpdateUserProfileParams{
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
		return outbox.InsertMessages(ctx, tx, schema, domainEvents)
	})
}

func rehydrateUser(row store.UsersUser) *domain.User {
	return domain.RehydrateUser(row.ID, row.Email, row.FirstName, row.LastName, row.IdentityID)
}
