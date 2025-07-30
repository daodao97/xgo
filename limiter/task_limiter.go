package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// 任务计数器

func NewTaskLimiter(windowHours int, maxRequests int, redisClient *redis.Client) *TaskLimiter {
	return &TaskLimiter{
		WindowHours: windowHours,
		MaxRequests: maxRequests,
		redisClient: redisClient,
	}
}

type TaskLimiter struct {
	WindowHours int
	MaxRequests int
	KeyPrefix   string
	redisClient *redis.Client
}

func (l *TaskLimiter) CanProcess(ctx context.Context, userID, resourceID string) bool {
	windowStart := time.Now().Add(-time.Duration(l.WindowHours) * time.Hour).Unix()
	taskCount, err := l.redisClient.ZCount(ctx, l.GetKey(userID, resourceID), fmt.Sprintf("%d", windowStart), "+inf").Result()
	if err != nil {
		return false
	}
	return int(taskCount) < l.MaxRequests
}

func (l *TaskLimiter) Process(ctx context.Context, userID, resourceID string) bool {
	if !l.CanProcess(ctx, userID, resourceID) {
		return false
	}
	l.redisClient.ZAdd(ctx, l.GetKey(userID, resourceID), &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: time.Now().UnixNano(),
	})
	return true
}

func (l *TaskLimiter) Finish(ctx context.Context, userID, resourceID string) {
	// No specific finish action needed for TaskLimiter
}

func (l *TaskLimiter) GetKey(userID, resourceID string) string {
	return fmt.Sprintf("%s:tasks:%s:%s", l.KeyPrefix, userID, resourceID)
}
