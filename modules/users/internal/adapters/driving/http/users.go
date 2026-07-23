package handler

import (
	"net/http"

	"github.com/llannillo/mm/internal/shared/auth"
	registeruser "github.com/llannillo/mm/modules/users/internal/app/commands/register_user"
	updateuser "github.com/llannillo/mm/modules/users/internal/app/commands/update_user"
	getuser "github.com/llannillo/mm/modules/users/internal/app/queries/get_user"
)

func (h *Handler) registerUser(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	req, ok := decodeJSON[request](w, r)
	if !ok {
		return
	}
	id, err := h.users.RegisterUser(r.Context(), registeruser.Command{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (h *Handler) getUserProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	resp, err := h.users.GetUser(r.Context(), getuser.Query{UserID: claims.UserID})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) updateUserProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	type request struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	req, ok := decodeJSON[request](w, r)
	if !ok {
		return
	}
	if err := h.users.UpdateUser(r.Context(), updateuser.Command{
		UserID:    claims.UserID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}); err != nil {
		handleDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
