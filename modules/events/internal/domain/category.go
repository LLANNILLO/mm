package domain

import "github.com/google/uuid"

var (
	ErrCategoryNotFound       = &DomainError{Code: "category.not_found", Message: "category not found", Kind: KindNotFound}
	ErrCategoryAlreadyArchived = &DomainError{Code: "category.already_archived", Message: "category is already archived", Kind: KindConflict}
	ErrCategoryNameEmpty      = &DomainError{Code: "category.name_empty", Message: "category name cannot be empty", Kind: KindValidation}
)

type Category struct {
	entity
	ID         uuid.UUID
	Name       string
	IsArchived bool
}

func NewCategory(name string) (*Category, error) {
	if name == "" {
		return nil, ErrCategoryNameEmpty
	}
	c := &Category{
		ID:         uuid.New(),
		Name:       name,
		IsArchived: false,
	}
	c.raise(CategoryCreatedDomainEvent{CategoryID: c.ID})
	return c, nil
}

func (c *Category) Archive() error {
	if c.IsArchived {
		return ErrCategoryAlreadyArchived
	}
	c.IsArchived = true
	c.raise(CategoryArchivedDomainEvent{CategoryID: c.ID})
	return nil
}

func (c *Category) ChangeName(name string) error {
	if name == "" {
		return ErrCategoryNameEmpty
	}
	if c.Name == name {
		return nil
	}
	c.Name = name
	c.raise(CategoryNameChangedDomainEvent{CategoryID: c.ID, Name: c.Name})
	return nil
}
