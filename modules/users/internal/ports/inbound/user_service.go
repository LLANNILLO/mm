package inbound

import (
	"context"

	"github.com/google/uuid"
	getuser "github.com/llannillo/mm/modules/users/internal/app/queries/get_user"
	registeruser "github.com/llannillo/mm/modules/users/internal/app/commands/register_user"
	updateuser "github.com/llannillo/mm/modules/users/internal/app/commands/update_user"
)

type UserService interface {
	RegisterUser(ctx context.Context, cmd registeruser.Command) (uuid.UUID, error)
	GetUser(ctx context.Context, q getuser.Query) (*getuser.Response, error)
	UpdateUser(ctx context.Context, cmd updateuser.Command) error
}
