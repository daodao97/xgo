package xredis

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type Counter struct {
	Prefix string
	Period time.Duration
}

func (r Counter) key(k string) string {
	return fmt.Sprintf("counter:%s:%s", r.Prefix, k)
}

func (r Counter) Incr(ctx context.Context, key string) bool {
	_key := r.key(key)
	val, err := client.Get(ctx, _key).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		return true
	}

	if val == 0 {
		client.Set(ctx, _key, 1, r.Period)
	} else {
		client.Incr(ctx, _key)
	}

	return true
}

func (r Counter) Get(ctx context.Context, key string) int64 {
	redisKey := r.key(key)
	i, err := client.Get(ctx, redisKey).Int64()
	if err != nil && !errors.Is(err, redis.Nil) {
		return 0
	}
	return i
}

// GetRemainingTime 获取剩余的限制时间
func (r Counter) GetRemainingTime(ctx context.Context, key string) time.Duration {
	redisKey := r.key(key)
	return client.TTL(ctx, redisKey).Val()
}
