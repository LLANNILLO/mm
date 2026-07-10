package domain

import "github.com/google/uuid"

var (
	ErrCategoryNotFound        = &DomainError{Code: "category.not_found", Message: "category not found", Kind: KindNotFound}
	ErrCategoryAlreadyArchived = &DomainError{Code: "category.already_archived", Message: "category is already archived", Kind: KindConflict}
	ErrCategoryNameEmpty       = &DomainError{Code: "category.name_empty", Message: "category name cannot be empty", Kind: KindValidation}
)

type Category struct {
	entity
	id         uuid.UUID
	name       string
	isArchived bool
}

func (c *Category) ID() uuid.UUID    { return c.id }
func (c *Category) Name() string     { return c.name }
func (c *Category) IsArchived() bool { return c.isArchived }

func NewCategory(name string) (*Category, error) {
	if name == "" {
		return nil, ErrCategoryNameEmpty
	}
	c := &Category{
		id:         uuid.New(),
		name:       name,
		isArchived: false,
	}
	c.raise(CategoryCreatedDomainEvent{CategoryID: c.id})
	return c, nil
}

// RehydrateCategory reconstructs a Category from persisted state without
// re-running creation invariants or raising domain events. Only repositories
// may call this.
func RehydrateCategory(id uuid.UUID, name string, isArchived bool) *Category {
	return &Category{id: id, name: name, isArchived: isArchived}
}

func (c *Category) Archive() error {
	if c.isArchived {
		return ErrCategoryAlreadyArchived
	}
	c.isArchived = true
	c.raise(CategoryArchivedDomainEvent{CategoryID: c.id})
	return nil
}

func (c *Category) ChangeName(name string) error {
	if name == "" {
		return ErrCategoryNameEmpty
	}
	if c.name == name {
		return nil
	}
	c.name = name
	c.raise(CategoryNameChangedDomainEvent{CategoryID: c.id, Name: c.name})
	return nil
}
