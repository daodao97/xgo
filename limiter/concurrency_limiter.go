package limiter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// 并发控制器

func NewConcurrencyLimiter(keyPrefix string, concurrentLimit int, redisClient redis.UniversalClient) *ConcurrencyLimiter {
	return &ConcurrencyLimiter{
		KeyPrefix:       keyPrefix,
		ConcurrentLimit: concurrentLimit,
		redisClient:     redisClient,
	}
}

type ConcurrencyLimiter struct {
	KeyPrefix       string
	ConcurrentLimit int
	redisClient     redis.UniversalClient
	TTL             time.Duration
}

func (l *ConcurrencyLimiter) CanProcess(ctx context.Context, userID, resourceID string) bool {
	currentConcurrent, err := l.redisClient.Get(ctx, l.GetKey(userID, resourceID)).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		return false
	}
	return currentConcurrent < l.ConcurrentLimit
}

func (l *ConcurrencyLimiter) Process(ctx context.Context, userID, resourceID string) bool {
	const script = `
local current = redis.call("GET", KEYS[1])
if not current then
	redis.call("SET", KEYS[1], 1)
	if tonumber(ARGV[2]) > 0 then
		redis.call("PEXPIRE", KEYS[1], ARGV[2])
	end
	return 1
end
if tonumber(current) < tonumber(ARGV[1]) then
	local v = redis.call("INCR", KEYS[1])
	if tonumber(ARGV[2]) > 0 then
		redis.call("PEXPIRE", KEYS[1], ARGV[2])
	end
	return v
end
return 0
`
	ttlMs := int64(l.TTL / time.Millisecond)
	res, err := l.redisClient.Eval(ctx, script, []string{l.GetKey(userID, resourceID)}, l.ConcurrentLimit, ttlMs).Int64()
	if err != nil {
		return false
	}
	return res > 0
}

func (l *ConcurrencyLimiter) Finish(ctx context.Context, userID, resourceID string) {
	const script = `
local current = redis.call("GET", KEYS[1])
if not current then
	return 0
end
local v = redis.call("DECR", KEYS[1])
if v <= 0 then
	redis.call("DEL", KEYS[1])
	return 0
end
return v
`
	_, _ = l.redisClient.Eval(ctx, script, []string{l.GetKey(userID, resourceID)}).Result()
}

func (l *ConcurrencyLimiter) GetKey(userID, resourceID string) string {
	return fmt.Sprintf("%s:concurrent:%s:%s", l.KeyPrefix, userID, resourceID)
}
