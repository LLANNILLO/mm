package handler

import (
	"net/http"

	createtickettype "github.com/llannillo/mm/modules/events/internal/app/commands/create_ticket_type"
	updateticketprice "github.com/llannillo/mm/modules/events/internal/app/commands/update_ticket_price"
	gettickettype "github.com/llannillo/mm/modules/events/internal/app/queries/get_ticket_type"
	listtickettype "github.com/llannillo/mm/modules/events/internal/app/queries/list_ticket_types"
)

func (h *Handler) createTicketType(w http.ResponseWriter, r *http.Request) {
	type request struct {
		EventID  string `json:"event_id"`
		Name     string `json:"name"`
		Price    int64  `json:"price"`
		Currency string `json:"currency"`
		Quantity int64  `json:"quantity"`
	}
	req, ok := decodeJSON[request](w, r)
	if !ok {
		return
	}
	eventID, ok := parseUUID(w, req.EventID)
	if !ok {
		return
	}
	id, err := h.tickets.CreateTicketType(r.Context(), createtickettype.Command{
		EventID:  eventID,
		Name:     req.Name,
		Price:    req.Price,
		Currency: req.Currency,
		Quantity: req.Quantity,
	})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (h *Handler) getTicketType(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("id"))
	if !ok {
		return
	}
	resp, err := h.tickets.GetTicketType(r.Context(), gettickettype.Query{ID: id})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) listTicketTypes(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseUUID(w, r.URL.Query().Get("event_id"))
	if !ok {
		return
	}
	items, err := h.tickets.ListTicketTypes(r.Context(), listtickettype.Query{EventID: eventID})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) updateTicketTypePrice(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("id"))
	if !ok {
		return
	}
	type request struct {
		Price int64 `json:"price"`
	}
	req, ok := decodeJSON[request](w, r)
	if !ok {
		return
	}
	if err := h.tickets.UpdateTicketPrice(r.Context(), updateticketprice.Command{
		TicketTypeID: id,
		Price:        req.Price,
	}); err != nil {
		handleDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
