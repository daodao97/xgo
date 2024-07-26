package cache

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("key not found in cache")
)

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
	Set(ctx context.Context, key string, data string) error
	SetWithTTL(ctx context.Context, key string, data string, ttl time.Duration) error
}
