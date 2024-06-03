package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// 任务计数器

type TaskLimiter struct {
	WindowHours int
	MaxRequests int
	redisClient *redis.Client
}

func (l *TaskLimiter) CanProcess(ctx context.Context, userID, resourceID string) bool {
	windowStart := time.Now().Add(-time.Duration(l.WindowHours) * time.Hour).Unix()
	taskCount, err := l.redisClient.ZCount(ctx, fmt.Sprintf("%s:tasks:%s", userID, resourceID), fmt.Sprintf("%d", windowStart), "+inf").Result()
	if err != nil {
		return false
	}
	return int(taskCount) < l.MaxRequests
}

func (l *TaskLimiter) Process(ctx context.Context, userID, resourceID string) bool {
	if !l.CanProcess(ctx, userID, resourceID) {
		return false
	}
	l.redisClient.ZAdd(ctx, fmt.Sprintf("%s:tasks:%s", userID, resourceID), &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: time.Now().UnixNano(),
	})
	return true
}

func (l *TaskLimiter) Finish(ctx context.Context, userID, resourceID string) {
	// No specific finish action needed for TaskLimiter
}
