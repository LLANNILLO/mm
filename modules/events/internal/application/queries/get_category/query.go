package getcategory

import (
	"github.com/google/uuid"
)

type Query struct {
	ID uuid.UUID
}

type Response struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	IsArchived bool      `json:"is_archived"`
}
