package cache

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisCache 是一个基于 Redis 的缓存实现
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache 创建一个新的 RedisCache 实例
func NewRedisCache(options *redis.Options) *RedisCache {
	// 创建一个 Redis 客户端
	client := redis.NewClient(options)

	// 检查是否能连接到 Redis 服务器
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		panic("failed to connect to Redis")
	}

	return &RedisCache{
		client: client,
	}
}

// NewRedis 创建一个新的 RedisCache 实例
func NewRedis(client *redis.Client) *RedisCache {
	// 检查是否能连接到 Redis 服务器
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		panic("failed to connect to Redis")
	}

	return &RedisCache{
		client: client,
	}
}

// Get 从 Redis 缓存中获取指定 key 的数据
func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrNotFound
	} else if err != nil {
		return "", err
	}
	return val, nil
}

// Del 从 Redis 缓存中删除指定 key 的数据
func (c *RedisCache) Del(ctx context.Context, key string) error {
	_, err := c.client.Del(ctx, key).Result()
	return err
}

// Set 将数据存储到 Redis 缓存中，如果 key 已存在，则覆盖原有数据
func (c *RedisCache) Set(ctx context.Context, key string, data string) error {
	err := c.client.Set(ctx, key, data, 0).Err()
	return err
}

// SetWithTTL 将数据存储到 Redis 缓存中，并设置存活时间（TTL）
func (c *RedisCache) SetWithTTL(ctx context.Context, key string, data string, ttl time.Duration) error {
	err := c.client.Set(ctx, key, data, ttl).Err()
	return err
}
