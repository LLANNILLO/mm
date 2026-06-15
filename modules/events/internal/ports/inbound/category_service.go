package inbound

import (
	"context"

	"github.com/google/uuid"
	archivecategory "github.com/llannillo/mm/modules/events/internal/app/commands/archive_category"
	createcategory "github.com/llannillo/mm/modules/events/internal/app/commands/create_category"
	renamecategory "github.com/llannillo/mm/modules/events/internal/app/commands/rename_category"
	getcategory "github.com/llannillo/mm/modules/events/internal/app/queries/get_category"
	listcategories "github.com/llannillo/mm/modules/events/internal/app/queries/list_categories"
)

type CategoryService interface {
	CreateCategory(ctx context.Context, cmd createcategory.Command) (uuid.UUID, error)
	ArchiveCategory(ctx context.Context, cmd archivecategory.Command) error
	RenameCategory(ctx context.Context, cmd renamecategory.Command) error
	GetCategory(ctx context.Context, q getcategory.Query) (*getcategory.Response, error)
	ListCategories(ctx context.Context) ([]listcategories.CategoryItem, error)
}
