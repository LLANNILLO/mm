package health

import (
	"context"

	"github.com/valkey-io/valkey-go"
)

type valkeyChecker struct {
	client valkey.Client
}

// NewValkeyChecker returns a Checker that sends PING to the Valkey server.
func NewValkeyChecker(client valkey.Client) Checker {
	return &valkeyChecker{client: client}
}

func (c *valkeyChecker) Check(ctx context.Context) error {
	return c.client.Do(ctx, c.client.B().Ping().Build()).Error()
}
