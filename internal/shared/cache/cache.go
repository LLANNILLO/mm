package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/valkey-io/valkey-go"
)

const defaultTTL = 2 * time.Minute

// ErrMiss is returned by Get when the key does not exist in the cache.
var ErrMiss = errors.New("cache miss")

// Service is the distributed cache abstraction.
type Service interface {
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, ttl ...time.Duration) error
	Remove(ctx context.Context, key string) error
}

type valkeyService struct {
	client valkey.Client
}

// NewService wraps a valkey.Client in the Service interface.
func NewService(client valkey.Client) Service {
	return &valkeyService{client: client}
}

func (s *valkeyService) Get(ctx context.Context, key string, dest any) error {
	b, err := s.client.Do(ctx, s.client.B().Get().Key(key).Build()).AsBytes()
	if err != nil {
		if valkey.IsValkeyNil(err) {
			return ErrMiss
		}
		return err
	}
	return json.Unmarshal(b, dest)
}

func (s *valkeyService) Set(ctx context.Context, key string, value any, ttl ...time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	expiry := defaultTTL
	if len(ttl) > 0 {
		expiry = ttl[0]
	}
	return s.client.Do(ctx, s.client.B().Set().Key(key).Value(valkey.BinaryString(b)).Ex(expiry).Build()).Error()
}

func (s *valkeyService) Remove(ctx context.Context, key string) error {
	return s.client.Do(ctx, s.client.B().Del().Key(key).Build()).Error()
}
