package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	store "github.com/llannillo/mm/modules/ticketing/internal/adapters/driven/postgres/generated"
	"github.com/llannillo/mm/modules/ticketing/internal/domain"
)

type CustomerRepository struct {
	queries *store.Queries
}

func NewCustomerRepository(q *store.Queries) *CustomerRepository {
	return &CustomerRepository{queries: q}
}

func (r *CustomerRepository) Insert(ctx context.Context, c *domain.Customer) error {
	err := r.queries.InsertCustomer(ctx, store.InsertCustomerParams{
		ID:        c.ID(),
		Email:     c.Email(),
		FirstName: c.FirstName(),
		LastName:  c.LastName(),
	})
	if err != nil {
		return fmt.Errorf("insert customer: %w", err)
	}
	return nil
}

func (r *CustomerRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	row, err := r.queries.SelectCustomerByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCustomerNotFound
		}
		return nil, fmt.Errorf("get customer: %w", err)
	}
	return domain.RehydrateCustomer(row.ID, row.Email, row.FirstName, row.LastName), nil
}

func (r *CustomerRepository) Update(ctx context.Context, c *domain.Customer) error {
	err := r.queries.UpdateCustomer(ctx, store.UpdateCustomerParams{
		ID:        c.ID(),
		FirstName: c.FirstName(),
		LastName:  c.LastName(),
	})
	if err != nil {
		return fmt.Errorf("update customer: %w", err)
	}
	return nil
}
