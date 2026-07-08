package auth

import (
	"context"

	"github.com/google/uuid"
)

type contextKey struct{}

type Claims struct {
	Sub         string
	Email       string
	UserID      uuid.UUID
	Permissions []string
}

func WithClaims(ctx context.Context, c Claims) context.Context {
	return context.WithValue(ctx, contextKey{}, c)
}

func ClaimsFromContext(ctx context.Context) (Claims, bool) {
	c, ok := ctx.Value(contextKey{}).(Claims)
	return c, ok
}
