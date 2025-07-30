package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/daodao97/xgo/limiter"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func main() {
	// 创建 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // 如果有密码请设置
		DB:       0,  // 默认数据库
	})

	// 测试 Redis 连接
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("无法连接到 Redis: %v", err)
	}

	// 创建滑动窗口限流器
	// 每个用户每分钟最多允许10个请求
	apiLimiter := limiter.NewSlidingWindowLimiter("api", 10, time.Minute, rdb)

	// 创建另一个限流器用于文件上传
	// 每个用户每小时最多允许5次文件上传
	uploadLimiter := limiter.NewSlidingWindowLimiter("upload", 5, time.Hour, rdb)

	// 创建 Gin 路由
	r := gin.Default()

	// API 限流中间件
	r.Use(func(c *gin.Context) {
		userID := getUserID(c) // 从请求中获取用户ID
		resourceID := "api"

		if !apiLimiter.Process(c.Request.Context(), userID, resourceID) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "请求过于频繁，请稍后再试",
				"limit": "每分钟最多10次请求",
			})
			c.Abort()
			return
		}

		c.Next()
	})

	// 普通API端点
	r.GET("/api/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "数据获取成功",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// 文件上传端点（有单独的限流）
	r.POST("/api/upload", func(c *gin.Context) {
		userID := getUserID(c)

		// 检查上传限流
		if !uploadLimiter.Process(c.Request.Context(), userID, "upload") {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "上传过于频繁，请稍后再试",
				"limit": "每小时最多5次上传",
			})
			return
		}

		// 模拟文件上传处理
		c.JSON(http.StatusOK, gin.H{
			"message": "文件上传成功",
		})
	})

	// 获取限流状态的端点
	r.GET("/api/limit-status", func(c *gin.Context) {
		userID := getUserID(c)

		// 获取当前窗口内的请求数量
		apiCount, _ := apiLimiter.GetCurrentWindowCount(c.Request.Context(), userID, "api")
		uploadCount, _ := uploadLimiter.GetCurrentWindowCount(c.Request.Context(), userID, "upload")

		c.JSON(http.StatusOK, gin.H{
			"api_requests_in_window":    apiCount,
			"api_limit":                 10,
			"upload_requests_in_window": uploadCount,
			"upload_limit":              5,
		})
	})

	// 启动服务器
	fmt.Println("服务器启动在 :8080")
	fmt.Println("测试端点:")
	fmt.Println("  GET  /api/data        - 普通API请求 (每分钟限制10次)")
	fmt.Println("  POST /api/upload      - 文件上传 (每小时限制5次)")
	fmt.Println("  GET  /api/limit-status - 查看限流状态")
	fmt.Println("\n可以使用 curl 进行测试:")
	fmt.Println("  curl http://localhost:8080/api/data")
	fmt.Println("  curl -X POST http://localhost:8080/api/upload")
	fmt.Println("  curl http://localhost:8080/api/limit-status")

	log.Fatal(http.ListenAndServe(":8080", r))
}

// 从请求中获取用户ID的辅助函数
func getUserID(c *gin.Context) string {
	// 在实际应用中，这里应该从JWT token、session或其他认证方式中获取用户ID
	// 这里为了演示，使用IP地址作为用户标识
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = c.ClientIP()
	}
	return userID
}
