package limiter

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// 并发控制器

type ConcurrencyLimiter struct {
	ConcurrentLimit int
	redisClient     *redis.Client
}

func (l *ConcurrencyLimiter) CanProcess(ctx context.Context, userID, resourceID string) bool {
	currentConcurrent, err := l.redisClient.Get(ctx, fmt.Sprintf("%s:concurrent:%s", userID, resourceID)).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		return false
	}
	return currentConcurrent < l.ConcurrentLimit
}

func (l *ConcurrencyLimiter) Process(ctx context.Context, userID, resourceID string) bool {
	if !l.CanProcess(ctx, userID, resourceID) {
		return false
	}
	l.redisClient.Incr(ctx, fmt.Sprintf("%s:concurrent:%s", userID, resourceID))
	return true
}

func (l *ConcurrencyLimiter) Finish(ctx context.Context, userID, resourceID string) {
	l.redisClient.Decr(ctx, fmt.Sprintf("%s:concurrent:%s", userID, resourceID))
}
