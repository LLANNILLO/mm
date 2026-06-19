package inbound

import (
	"context"

	additemtocart "github.com/llannillo/mm/modules/ticketing/internal/app/commands/add_item_to_cart"
)

type CartService interface {
	AddItemToCart(ctx context.Context, cmd additemtocart.Command) error
}
