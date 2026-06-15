package presentation

import (
	"net/http"

	archivecategory "github.com/llannillo/mm/modules/events/internal/application/commands/archive_category"
	createcategory "github.com/llannillo/mm/modules/events/internal/application/commands/create_category"
	renamecategory "github.com/llannillo/mm/modules/events/internal/application/commands/rename_category"
	getcategory "github.com/llannillo/mm/modules/events/internal/application/queries/get_category"
)

func (h *Handler) createCategory(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Name string `json:"name"`
	}
	req, ok := decodeJSON[request](w, r)
	if !ok {
		return
	}
	id, err := h.deps.CreateCategory.Handle(r.Context(), createcategory.Command{Name: req.Name})
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
	resp, err := h.deps.GetCategory.Handle(r.Context(), getcategory.Query{ID: id})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	items, err := h.deps.ListCategories.Handle(r.Context())
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
	if err := h.deps.ArchiveCategory.Handle(r.Context(), archivecategory.Command{CategoryID: id}); err != nil {
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
	if err := h.deps.RenameCategory.Handle(r.Context(), renamecategory.Command{
		CategoryID: id,
		Name:       req.Name,
	}); err != nil {
		handleDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
