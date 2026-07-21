package http

import (
	"net/http"

	"github.com/google/uuid"
	createorder "github.com/llannillo/mm/modules/ticketing/internal/app/commands/create_order"
)

type createOrderRequest struct {
	CustomerID  uuid.UUID          `json:"customer_id"`
	TicketTypes []orderItemRequest `json:"ticket_types"`
}

type orderItemRequest struct {
	TicketTypeID uuid.UUID `json:"ticket_type_id"`
	Quantity     int64     `json:"quantity"`
}

func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[createOrderRequest](w, r)
	if !ok {
		return
	}

	items := make([]createorder.OrderItem, 0, len(req.TicketTypes))
	for _, i := range req.TicketTypes {
		items = append(items, createorder.OrderItem{
			TicketTypeID: i.TicketTypeID,
			Quantity:     i.Quantity,
		})
	}

	id, err := h.orders.CreateOrder(r.Context(), createorder.Command{
		CustomerID:  req.CustomerID,
		TicketTypes: items,
	})
	if err != nil {
		handleDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}
