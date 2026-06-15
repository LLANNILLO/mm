package handler

import (
	"net/http"

	archivecategory "github.com/llannillo/mm/modules/events/internal/app/commands/archive_category"
	createcategory "github.com/llannillo/mm/modules/events/internal/app/commands/create_category"
	renamecategory "github.com/llannillo/mm/modules/events/internal/app/commands/rename_category"
	getcategory "github.com/llannillo/mm/modules/events/internal/app/queries/get_category"
)

func (h *Handler) createCategory(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Name string `json:"name"`
	}
	req, ok := decodeJSON[request](w, r)
	if !ok {
		return
	}
	id, err := h.categories.CreateCategory(r.Context(), createcategory.Command{Name: req.Name})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (h *Handler) getCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("id"))
	if !ok {
		return
	}
	resp, err := h.categories.GetCategory(r.Context(), getcategory.Query{ID: id})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	items, err := h.categories.ListCategories(r.Context())
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) archiveCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("id"))
	if !ok {
		return
	}
	if err := h.categories.ArchiveCategory(r.Context(), archivecategory.Command{CategoryID: id}); err != nil {
		handleDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) renameCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("id"))
	if !ok {
		return
	}
	type request struct {
		Name string `json:"name"`
	}
	req, ok := decodeJSON[request](w, r)
	if !ok {
		return
	}
	if err := h.categories.RenameCategory(r.Context(), renamecategory.Command{
		CategoryID: id,
		Name:       req.Name,
	}); err != nil {
		handleDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
