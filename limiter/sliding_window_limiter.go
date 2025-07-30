package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// SlidingWindowLimiter 基于时间滑动窗口的限流器
type SlidingWindowLimiter struct {
	KeyPrefix   string
	Limit       int           // 时间窗口内允许的最大请求数
	WindowSize  time.Duration // 时间窗口大小
	redisClient *redis.Client
}

// NewSlidingWindowLimiter 创建一个新的滑动窗口限流器
func NewSlidingWindowLimiter(keyPrefix string, limit int, windowSize time.Duration, redisClient *redis.Client) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		KeyPrefix:   keyPrefix,
		Limit:       limit,
		WindowSize:  windowSize,
		redisClient: redisClient,
	}
}

// CanProcess 检查是否可以处理请求
func (l *SlidingWindowLimiter) CanProcess(ctx context.Context, userID, resourceID string) bool {
	key := l.GetKey(userID, resourceID)
	now := time.Now()
	windowStart := now.Add(-l.WindowSize)

	// 清理过期的记录
	l.redisClient.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixNano()))

	// 获取当前窗口内的请求数量
	count, err := l.redisClient.ZCard(ctx, key).Result()
	if err != nil {
		return false
	}

	return count < int64(l.Limit)
}

// Process 处理请求，如果允许则记录请求时间戳
func (l *SlidingWindowLimiter) Process(ctx context.Context, userID, resourceID string) bool {
	if !l.CanProcess(ctx, userID, resourceID) {
		return false
	}

	key := l.GetKey(userID, resourceID)
	now := time.Now()

	// 添加当前请求的时间戳到有序集合
	_, err := l.redisClient.ZAdd(ctx, key, &redis.Z{
		Score:  float64(now.UnixNano()),
		Member: now.UnixNano(),
	}).Result()

	if err != nil {
		return false
	}

	// 设置键的过期时间，防止内存泄漏
	l.redisClient.Expire(ctx, key, l.WindowSize*2)

	return true
}

// Finish 完成请求处理
// 对于滑动窗口限流器，不需要在请求完成时做特殊处理，请求记录会自动过期
func (l *SlidingWindowLimiter) Finish(ctx context.Context, userID, resourceID string) {
	// 对于滑动窗口限流器，请求记录会通过时间窗口自动清理
	// 这里可以选择性地进行一些清理操作，但通常不需要
}

// GetKey 生成Redis键名
func (l *SlidingWindowLimiter) GetKey(userID, resourceID string) string {
	return fmt.Sprintf("%s:sliding_window:%s:%s", l.KeyPrefix, userID, resourceID)
}

// GetCurrentWindowCount 获取当前时间窗口内的请求数量（可选的辅助方法）
func (l *SlidingWindowLimiter) GetCurrentWindowCount(ctx context.Context, userID, resourceID string) (int64, error) {
	key := l.GetKey(userID, resourceID)
	now := time.Now()
	windowStart := now.Add(-l.WindowSize)

	// 清理过期的记录
	l.redisClient.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixNano()))

	// 获取当前窗口内的请求数量
	return l.redisClient.ZCard(ctx, key).Result()
}
