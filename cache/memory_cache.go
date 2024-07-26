package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

// MemoryCache 是一个基于内存的缓存实现
type MemoryCache struct {
	cache    map[string]string
	mutex    sync.RWMutex
	timers   map[string]*time.Timer
	timersMu sync.Mutex
}

// NewMemoryCache 创建一个新的 MemoryCache 实例
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		cache:  make(map[string]string),
		timers: make(map[string]*time.Timer),
	}
}

// Get 从缓存中获取指定 key 的数据
func (c *MemoryCache) Get(ctx context.Context, key string) (string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	data, ok := c.cache[key]
	if !ok {
		return "", ErrNotFound
	}
	return data, nil
}

// Del 从缓存中删除指定 key 的数据
func (c *MemoryCache) Del(ctx context.Context, key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	_, ok := c.cache[key]
	if !ok {
		return errors.New("key not found in cache")
	}
	delete(c.cache, key)
	return nil
}

// Set 将数据存储到缓存中，如果 key 已存在，则覆盖原有数据
func (c *MemoryCache) Set(ctx context.Context, key string, data string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[key] = data
	return nil
}

// SetWithTTL 将数据存储到缓存中，并设置存活时间（TTL）
func (c *MemoryCache) SetWithTTL(ctx context.Context, key string, data string, ttl time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 如果存在之前的定时器，取消它
	c.timersMu.Lock()
	if oldTimer, ok := c.timers[key]; ok {
		oldTimer.Stop()
		delete(c.timers, key)
	}
	c.timersMu.Unlock()

	c.cache[key] = data

	// 设置新的定时器
	timer := time.AfterFunc(ttl, func() {
		c.mutex.Lock()
		defer c.mutex.Unlock()
		delete(c.cache, key)
	})

	c.timersMu.Lock()
	c.timers[key] = timer
	c.timersMu.Unlock()

	return nil
}
