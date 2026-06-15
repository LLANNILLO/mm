package listcategories

import "github.com/google/uuid"

type CategoryItem struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	IsArchived bool      `json:"is_archived"`
}
