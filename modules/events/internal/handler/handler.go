package handler

import "github.com/llannillo/mm/modules/events/internal/store"

type Handler struct {
	queries *store.Queries
}

func New(q *store.Queries) *Handler {
	return &Handler{
		queries: q,
	}
}
