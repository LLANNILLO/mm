package http

import (
	"net/http"

	"github.com/google/uuid"
	additemtocart "github.com/llannillo/mm/modules/ticketing/internal/app/commands/add_item_to_cart"
)

type addToCartRequest struct {
	CustomerID   uuid.UUID `json:"customer_id"`
	TicketTypeID uuid.UUID `json:"ticket_type_id"`
	Quantity     int64     `json:"quantity"`
}

func (h *Handler) addToCart(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[addToCartRequest](w, r)
	if !ok {
		return
	}

	err := h.carts.AddItemToCart(r.Context(), additemtocart.Command{
		CustomerID:   req.CustomerID,
		TicketTypeID: req.TicketTypeID,
		Quantity:     req.Quantity,
	})
	if err != nil {
		handleDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
