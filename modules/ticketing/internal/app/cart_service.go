package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/llannillo/mm/internal/shared/cache"
	"github.com/llannillo/mm/modules/ticketing/internal/domain"
)

const cartTTL = 20 * time.Minute

type CartService struct {
	cache cache.Service
}

func NewCartService(c cache.Service) *CartService {
	return &CartService{cache: c}
}

func (s *CartService) GetCart(ctx context.Context, customerID uuid.UUID) (*domain.Cart, error) {
	var cart domain.Cart
	err := s.cache.Get(ctx, cartKey(customerID), &cart)
	if errors.Is(err, cache.ErrMiss) {
		return &domain.Cart{CustomerID: customerID}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get cart: %w", err)
	}
	return &cart, nil
}

func (s *CartService) AddItem(ctx context.Context, customerID uuid.UUID, item domain.CartItem) error {
	cart, err := s.GetCart(ctx, customerID)
	if err != nil {
		return err
	}

	for i, existing := range cart.Items {
		if existing.TicketTypeID == item.TicketTypeID {
			cart.Items[i].Quantity += item.Quantity
			return s.cache.Set(ctx, cartKey(customerID), cart, cartTTL)
		}
	}

	cart.Items = append(cart.Items, item)
	return s.cache.Set(ctx, cartKey(customerID), cart, cartTTL)
}

func (s *CartService) ClearCart(ctx context.Context, customerID uuid.UUID) error {
	return s.cache.Remove(ctx, cartKey(customerID))
}

func cartKey(customerID uuid.UUID) string {
	return "carts:" + customerID.String()
}
