package limiter

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestRunner(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Redis server address
	})

	// Define different limiters based on user type
	concurrencyLimiter := &ConcurrencyLimiter{
		ConcurrentLimit: 1,
		redisClient:     rdb,
	}

	taskLimiter := &TaskLimiter{
		WindowHours: 24,
		MaxRequests: 50,
		redisClient: rdb,
	}

	modelLimiter := &ModelLimiter{
		AllowedResources: map[string]bool{
			"slow":  true,
			"fast":  false,
			"ultra": false,
		},
	}

	// Create LimiterRunner for a Free user
	limiterRunner := NewRunner(modelLimiter, taskLimiter, concurrencyLimiter)

	userID := "user1"
	resourceID := "slow"

	ctx := context.Background()

	if limiterRunner.Process(ctx, userID, resourceID) {
		fmt.Println("Request processed")
		// Simulate request handling
		time.Sleep(2 * time.Second)
		limiterRunner.Finish(ctx, userID, resourceID)
	} else {
		fmt.Println("Request denied")
	}
}
