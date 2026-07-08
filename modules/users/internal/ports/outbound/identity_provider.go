package outbound

import "context"

type IdentityProvider interface {
	RegisterUser(ctx context.Context, email, password, firstName, lastName string) (identityID string, err error)
}
