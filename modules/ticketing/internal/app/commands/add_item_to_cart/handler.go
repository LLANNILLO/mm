package additemtocart

import (
	"context"
	"fmt"

	eventsapi "github.com/llannillo/mm/modules/events/api"
	usersapi "github.com/llannillo/mm/modules/users/api"
	"github.com/llannillo/mm/modules/ticketing/internal/app"
	"github.com/llannillo/mm/modules/ticketing/internal/domain"
)

type Handler struct {
	cartService *app.CartService
	eventsAPI   eventsapi.EventsAPI
	usersAPI    usersapi.UsersAPI
}

func NewHandler(cartService *app.CartService, eventsAPI eventsapi.EventsAPI, usersAPI usersapi.UsersAPI) *Handler {
	return &Handler{
		cartService: cartService,
		eventsAPI:   eventsAPI,
		usersAPI:    usersAPI,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	customer, err := h.usersAPI.GetUser(ctx, cmd.CustomerID)
	if err != nil {
		return fmt.Errorf("get customer: %w", err)
	}
	if customer == nil {
		return domain.ErrCustomerNotFound
	}

	ticketType, err := h.eventsAPI.GetTicketType(ctx, cmd.TicketTypeID)
	if err != nil {
		return fmt.Errorf("get ticket type: %w", err)
	}
	if ticketType == nil {
		return domain.ErrTicketTypeNotFound
	}

	item := domain.CartItem{
		TicketTypeID: ticketType.ID,
		Quantity:     cmd.Quantity,
		Price:        ticketType.Price,
		Currency:     ticketType.Currency,
	}

	if err := h.cartService.AddItem(ctx, cmd.CustomerID, item); err != nil {
		return fmt.Errorf("add item to cart: %w", err)
	}

	return nil
}
