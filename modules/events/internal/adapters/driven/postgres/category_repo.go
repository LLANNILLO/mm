package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/llannillo/mm/modules/events/internal/domain"
	store "github.com/llannillo/mm/modules/events/internal/adapters/driven/postgres/generated"
)

type CategoryRepository struct {
	queries *store.Queries
}

func NewCategoryRepository(q *store.Queries) *CategoryRepository {
	return &CategoryRepository{queries: q}
}

func (r *CategoryRepository) Insert(ctx context.Context, category *domain.Category) error {
	_, err := r.queries.InsertCategory(ctx, store.InsertCategoryParams{
		ID:   category.ID,
		Name: category.Name,
	})
	if err != nil {
		return fmt.Errorf("insert category: %w", err)
	}
	return nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	row, err := r.queries.SelectCategoryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("get category by id: %w", err)
	}
	return &domain.Category{
		ID:         row.ID,
		Name:       row.Name,
		IsArchived: row.IsArchived,
	}, nil
}

func (r *CategoryRepository) Update(ctx context.Context, c *domain.Category) error {
	for _, e := range c.DomainEvents() {
		switch e.(type) {
		case domain.CategoryArchivedDomainEvent:
			return r.queries.UpdateCategoryArchived(ctx, c.ID)
		case domain.CategoryNameChangedDomainEvent:
			return r.queries.UpdateCategoryName(ctx, store.UpdateCategoryNameParams{
				ID:   c.ID,
				Name: c.Name,
			})
		}
	}
	return nil
}
