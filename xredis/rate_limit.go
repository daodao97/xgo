package xredis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

func NewRateLimiter(period time.Duration, limit int, prefix string) *RateLimit {
	return &RateLimit{
		Period: period,
		Limit:  limit,
		Prefix: prefix,
	}
}

type RateLimit struct {
	Period time.Duration
	Limit  int
	Prefix string
}

// IsAllowed 检查是否允许进行下一次API调用
func (r RateLimit) IsAllowed(ctx context.Context, key string) bool {
	redisKey := fmt.Sprintf("rate_limit_%s:%s", r.Prefix, key)
	val, err := client.Get(ctx, redisKey).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		return true
	}

	return val < r.Limit
}

func (r RateLimit) Incr(ctx context.Context, key string) bool {
	redisKey := fmt.Sprintf("rate_limit_%s:%s", r.Prefix, key)
	val, err := client.Get(ctx, redisKey).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		return true
	}

	if val == 0 {
		client.Set(ctx, redisKey, 1, r.Period)
	} else {
		client.Incr(ctx, redisKey)
	}

	return true
}

// GetRemainingTime 获取剩余的限制时间
func (r RateLimit) GetRemainingTime(ctx context.Context, key string) time.Duration {
	redisKey := fmt.Sprintf("rate_limit_%s:%s:%d", r.Prefix, key, time.Now().UnixNano()/int64(r.Period))
	return client.TTL(ctx, redisKey).Val()
}