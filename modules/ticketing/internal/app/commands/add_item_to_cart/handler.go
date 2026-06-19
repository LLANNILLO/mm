package additemtocart

import (
	"context"
	"fmt"

	"github.com/llannillo/mm/modules/ticketing/internal/app"
	"github.com/llannillo/mm/modules/ticketing/internal/domain"
	"github.com/llannillo/mm/modules/ticketing/internal/ports/outbound"
)

type Handler struct {
	cartService    *app.CartService
	customerRepo   outbound.CustomerRepository
	ticketTypeRepo outbound.TicketTypeRepository
}

func NewHandler(cartService *app.CartService, customerRepo outbound.CustomerRepository, ticketTypeRepo outbound.TicketTypeRepository) *Handler {
	return &Handler{
		cartService:    cartService,
		customerRepo:   customerRepo,
		ticketTypeRepo: ticketTypeRepo,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	_, err := h.customerRepo.GetByID(ctx, cmd.CustomerID)
	if err != nil {
		return fmt.Errorf("get customer: %w", err)
	}

	ticketType, err := h.ticketTypeRepo.GetByID(ctx, cmd.TicketTypeID)
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
