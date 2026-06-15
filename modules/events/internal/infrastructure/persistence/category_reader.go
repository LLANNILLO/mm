package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	getcategory "github.com/llannillo/mm/modules/events/internal/application/queries/get_category"
	listcategories "github.com/llannillo/mm/modules/events/internal/application/queries/list_categories"
	"github.com/llannillo/mm/modules/events/internal/domain"
	store "github.com/llannillo/mm/modules/events/internal/infrastructure/store/generated"
)

type CategoryReader struct {
	queries *store.Queries
}

func NewCategoryReader(q *store.Queries) *CategoryReader {
	return &CategoryReader{queries: q}
}

func (r *CategoryReader) GetCategory(ctx context.Context, id uuid.UUID) (*getcategory.Response, error) {
	row, err := r.queries.SelectCategoryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("get category: %w", err)
	}
	return &getcategory.Response{
		ID:         row.ID,
		Name:       row.Name,
		IsArchived: row.IsArchived,
	}, nil
}

func (r *CategoryReader) ListCategories(ctx context.Context) ([]listcategories.CategoryItem, error) {
	rows, err := r.queries.SelectCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	items := make([]listcategories.CategoryItem, len(rows))
	for i, row := range rows {
		items[i] = listcategories.CategoryItem{
			ID:         row.ID,
			Name:       row.Name,
			IsArchived: row.IsArchived,
		}
	}
	return items, nil
}
