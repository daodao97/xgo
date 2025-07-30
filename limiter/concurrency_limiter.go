package limiter

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// 并发控制器

func NewConcurrencyLimiter(keyPrefix string, concurrentLimit int, redisClient *redis.Client) *ConcurrencyLimiter {
	return &ConcurrencyLimiter{
		KeyPrefix:       keyPrefix,
		ConcurrentLimit: concurrentLimit,
		redisClient:     redisClient,
	}
}

type ConcurrencyLimiter struct {
	KeyPrefix       string
	ConcurrentLimit int
	redisClient     *redis.Client
}

func (l *ConcurrencyLimiter) CanProcess(ctx context.Context, userID, resourceID string) bool {
	currentConcurrent, err := l.redisClient.Get(ctx, l.GetKey(userID, resourceID)).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		return false
	}
	return currentConcurrent < l.ConcurrentLimit
}

func (l *ConcurrencyLimiter) Process(ctx context.Context, userID, resourceID string) bool {
	if !l.CanProcess(ctx, userID, resourceID) {
		return false
	}
	l.redisClient.Incr(ctx, l.GetKey(userID, resourceID))
	return true
}

func (l *ConcurrencyLimiter) Finish(ctx context.Context, userID, resourceID string) {
	l.redisClient.Decr(ctx, l.GetKey(userID, resourceID))
}

func (l *ConcurrencyLimiter) GetKey(userID, resourceID string) string {
	return fmt.Sprintf("%s:concurrent:%s:%s", l.KeyPrefix, userID, resourceID)
}
