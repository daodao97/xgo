package cache

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisOption func(c *RedisCache)

func WithPrefix(prefix string) RedisOption {
	return func(c *RedisCache) {
		c.prefix = prefix
	}
}

// RedisCache 是一个基于 Redis 的缓存实现
type RedisCache struct {
	client redis.UniversalClient
	prefix string
}

// NewRedisCache 创建一个新的 RedisCache 实例
func NewRedisCache(options *redis.Options, option ...RedisOption) *RedisCache {
	// 创建一个 Redis 客户端
	client := redis.NewClient(options)

	// 检查是否能连接到 Redis 服务器
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		panic("failed to connect to Redis")
	}

	c := &RedisCache{
		client: client,
	}

	for _, o := range option {
		o(c)
	}
	return c
}

// NewRedis 创建一个新的 RedisCache 实例
func NewRedis(client redis.UniversalClient, option ...RedisOption) *RedisCache {
	// 检查是否能连接到 Redis 服务器
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		panic("failed to connect to Redis")
	}

	c := &RedisCache{
		client: client,
	}
	for _, o := range option {
		o(c)
	}
	return c
}

func (c *RedisCache) key(key string) string {
	if c.prefix != "" {
		return c.prefix + ":" + key
	}
	return key
}

// Get 从 Redis 缓存中获取指定 key 的数据
func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, c.key(key)).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrNotFound
	} else if err != nil {
		return "", err
	}
	return val, nil
}

// Del 从 Redis 缓存中删除指定 key 的数据
func (c *RedisCache) Del(ctx context.Context, key string) error {
	_, err := c.client.Del(ctx, c.key(key)).Result()
	return err
}

// Set 将数据存储到 Redis 缓存中，如果 key 已存在，则覆盖原有数据
func (c *RedisCache) Set(ctx context.Context, key string, data string) error {
	err := c.client.Set(ctx, c.key(key), data, 0).Err()
	return err
}

// SetWithTTL 将数据存储到 Redis 缓存中，并设置存活时间（TTL）
func (c *RedisCache) SetWithTTL(ctx context.Context, key string, data string, ttl time.Duration) error {
	err := c.client.Set(ctx, c.key(key), data, ttl).Err()
	return err
}
