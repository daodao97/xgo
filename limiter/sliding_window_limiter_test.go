package limiter

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestSlidingWindowLimiter_Interface(t *testing.T) {
	// 确保 SlidingWindowLimiter 实现了 Limiter 接口
	var _ Limiter = (*SlidingWindowLimiter)(nil)
}

func TestSlidingWindowLimiter_BasicFunctionality(t *testing.T) {
	// 这里使用模拟的 Redis 客户端进行测试
	// 在实际项目中，您可能需要使用 testcontainers 或 miniredis 来进行集成测试

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // 使用测试数据库
	})

	// 检查 Redis 连接
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis not available, skipping integration test")
		return
	}

	// 清理测试数据
	defer func() {
		keys, _ := rdb.Keys(ctx, "test:sliding_window:*").Result()
		if len(keys) > 0 {
			rdb.Del(ctx, keys...)
		}
		rdb.Close()
	}()

	// 创建限流器：1秒内最多5个请求
	limiter := NewSlidingWindowLimiter("test", 5, time.Second, rdb)

	userID := "user1"
	resourceID := "resource1"

	// 测试初始状态
	assert.True(t, limiter.CanProcess(ctx, userID, resourceID))

	// 测试正常处理
	for i := 0; i < 5; i++ {
		assert.True(t, limiter.Process(ctx, userID, resourceID))
	}

	// 第6个请求应该被拒绝
	assert.False(t, limiter.CanProcess(ctx, userID, resourceID))
	assert.False(t, limiter.Process(ctx, userID, resourceID))

	// 等待一段时间后应该可以继续处理
	time.Sleep(1100 * time.Millisecond)
	assert.True(t, limiter.CanProcess(ctx, userID, resourceID))
	assert.True(t, limiter.Process(ctx, userID, resourceID))
}

func TestSlidingWindowLimiter_GetCurrentWindowCount(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis not available, skipping integration test")
		return
	}

	defer func() {
		keys, _ := rdb.Keys(ctx, "test:sliding_window:*").Result()
		if len(keys) > 0 {
			rdb.Del(ctx, keys...)
		}
		rdb.Close()
	}()

	limiter := NewSlidingWindowLimiter("test", 10, time.Second, rdb)
	userID := "user2"
	resourceID := "resource2"

	// 初始计数应该为0
	count, err := limiter.GetCurrentWindowCount(ctx, userID, resourceID)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// 处理3个请求
	for i := 0; i < 3; i++ {
		assert.True(t, limiter.Process(ctx, userID, resourceID))
	}

	// 计数应该为3
	count, err = limiter.GetCurrentWindowCount(ctx, userID, resourceID)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestSlidingWindowLimiter_MultipleUsers(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis not available, skipping integration test")
		return
	}

	defer func() {
		keys, _ := rdb.Keys(ctx, "test:sliding_window:*").Result()
		if len(keys) > 0 {
			rdb.Del(ctx, keys...)
		}
		rdb.Close()
	}()

	limiter := NewSlidingWindowLimiter("test", 2, time.Second, rdb)

	user1 := "user1"
	user2 := "user2"
	resourceID := "resource1"

	// 用户1处理2个请求
	assert.True(t, limiter.Process(ctx, user1, resourceID))
	assert.True(t, limiter.Process(ctx, user1, resourceID))
	assert.False(t, limiter.Process(ctx, user1, resourceID)) // 第3个被拒绝

	// 用户2应该仍然可以处理请求
	assert.True(t, limiter.Process(ctx, user2, resourceID))
	assert.True(t, limiter.Process(ctx, user2, resourceID))
	assert.False(t, limiter.Process(ctx, user2, resourceID)) // 第3个被拒绝
}

// 基准测试
func BenchmarkSlidingWindowLimiter_Process(b *testing.B) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		b.Skip("Redis not available, skipping benchmark")
		return
	}

	defer rdb.Close()

	limiter := NewSlidingWindowLimiter("benchmark", 1000000, time.Minute, rdb)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		userID := "benchuser"
		resourceID := "benchresource"
		i := 0
		for pb.Next() {
			limiter.Process(ctx, userID, resourceID)
			i++
		}
	})
}
