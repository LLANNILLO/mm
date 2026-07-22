package http

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/llannillo/mm/modules/ticketing/internal/ports/inbound"
)

type Handler struct {
	carts   inbound.CartService
	orders  inbound.OrderService
	tickets inbound.TicketService
}

func NewHandler(carts inbound.CartService, orders inbound.OrderService, tickets inbound.TicketService) *Handler {
	return &Handler{carts: carts, orders: orders, tickets: tickets}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("PUT /carts/add", h.addToCart)
	mux.HandleFunc("POST /orders", h.createOrder)
	mux.HandleFunc("POST /tickets/{id}/check-in", h.checkInTicket)
	mux.HandleFunc("GET /events/{id}/statistics", h.getEventStatistics)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decodeJSON[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return v, false
	}
	return v, true
}

func parseUUID(w http.ResponseWriter, raw string) (uuid.UUID, bool) {
	id, err := uuid.Parse(raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return uuid.Nil, false
	}
	return id, true
}
