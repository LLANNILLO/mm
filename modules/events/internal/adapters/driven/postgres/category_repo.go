package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/llannillo/mm/internal/shared/outbox"
	store "github.com/llannillo/mm/modules/events/internal/adapters/driven/postgres/generated"
	"github.com/llannillo/mm/modules/events/internal/domain"
)

type CategoryRepository struct {
	queries *store.Queries
	uow     *UnitOfWork
}

func NewCategoryRepository(q *store.Queries, uow *UnitOfWork) *CategoryRepository {
	return &CategoryRepository{queries: q, uow: uow}
}

func (r *CategoryRepository) Insert(ctx context.Context, category *domain.Category) error {
	return r.uow.WithTx(ctx, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		_, err := q.InsertCategory(ctx, store.InsertCategoryParams{
			ID:   category.ID(),
			Name: category.Name(),
		})
		if err != nil {
			return fmt.Errorf("insert category: %w", err)
		}

		domainEvents := category.DomainEvents()
		category.ClearDomainEvents()
		return outbox.InsertMessages(ctx, tx, schema, domainEvents)
	})
}

func (r *CategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	row, err := r.queries.SelectCategoryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("get category by id: %w", err)
	}
	return domain.RehydrateCategory(row.ID, row.Name, row.IsArchived), nil
}

func (r *CategoryRepository) Update(ctx context.Context, c *domain.Category) error {
	return r.uow.WithTx(ctx, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		for _, e := range c.DomainEvents() {
			var err error
			switch e.(type) {
			case domain.CategoryArchivedDomainEvent:
				err = q.UpdateCategoryArchived(ctx, c.ID())
			case domain.CategoryNameChangedDomainEvent:
				err = q.UpdateCategoryName(ctx, store.UpdateCategoryNameParams{
					ID:   c.ID(),
					Name: c.Name(),
				})
			}
			if err != nil {
				return err
			}
		}

		domainEvents := c.DomainEvents()
		c.ClearDomainEvents()
		return outbox.InsertMessages(ctx, tx, schema, domainEvents)
	})
}
