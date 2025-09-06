package limiter

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRunner(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Redis server address
	})

	// Define different limiters based on user type
	concurrencyLimiter := NewConcurrencyLimiter("test", 1, rdb)

	taskLimiter := NewTaskLimiter(24, 50, rdb)

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
