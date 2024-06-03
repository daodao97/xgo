package xredis

import (
	"context"

	"github.com/go-redis/redis/v8"
)

var client *redis.Client

func Init(opt *redis.Options) error {
	client = redis.NewClient(opt)
	return client.Ping(context.Background()).Err()
}

func Get() *redis.Client {
	return client
}
