package http

import (
	"net/http"

	"github.com/google/uuid"
	checkinticket "github.com/llannillo/mm/modules/ticketing/internal/app/commands/check_in_ticket"
	geteventstatistics "github.com/llannillo/mm/modules/ticketing/internal/app/queries/get_event_statistics"
)

type checkInRequest struct {
	CustomerID uuid.UUID `json:"customer_id"`
}

func (h *Handler) checkInTicket(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("id"))
	if !ok {
		return
	}
	req, ok := decodeJSON[checkInRequest](w, r)
	if !ok {
		return
	}

	if err := h.tickets.CheckIn(r.Context(), checkinticket.Command{
		TicketID:   id,
		CustomerID: req.CustomerID,
	}); err != nil {
		handleDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getEventStatistics(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("id"))
	if !ok {
		return
	}

	resp, err := h.tickets.GetEventStatistics(r.Context(), geteventstatistics.Query{EventID: id})
	if err != nil {
		handleDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
