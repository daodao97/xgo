package xredis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimitOptions struct {
	Messags string
}

type RateLimitOption func(options *RateLimitOptions)

func NewRateLimiter(period time.Duration, limit int, prefix string, opts ...RateLimitOption) *RateLimit {
	options := &RateLimitOptions{}
	for _, opt := range opts {
		opt(options)
	}

	return &RateLimit{
		Period:  period,
		Limit:   limit,
		Prefix:  prefix,
		Messags: options.Messags,
	}
}

type RateLimit struct {
	Period  time.Duration
	Limit   int
	Prefix  string
	Messags string
}

func (r RateLimit) key(k string) string {
	return fmt.Sprintf("rate_limit_%s:%s", r.Prefix, k)
}

// IsAllowed 检查是否允许进行下一次API调用
func (r RateLimit) IsAllowed(ctx context.Context, key string) bool {
	redisKey := r.key(key)
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

func (r RateLimit) IncrBy(ctx context.Context, key string, value int) bool {
	redisKey := fmt.Sprintf("rate_limit_%s:%s", r.Prefix, key)
	val, err := client.Get(ctx, redisKey).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		return true
	}

	if val == 0 {
		client.Set(ctx, redisKey, value, r.Period)
	} else {
		client.IncrBy(ctx, redisKey, int64(value))
	}

	return true
}

func (r RateLimit) Get(ctx context.Context, key string) int64 {
	redisKey := r.key(key)
	i, err := client.Get(ctx, redisKey).Int64()
	if err != nil && !errors.Is(err, redis.Nil) {
		return 0
	}
	return i
}

// GetRemainingTime 获取剩余的限制时间
func (r RateLimit) GetRemainingTime(ctx context.Context, key string) time.Duration {
	redisKey := r.key(key)
	return client.TTL(ctx, redisKey).Val()
}
