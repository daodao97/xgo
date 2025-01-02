package xredis

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

type Options struct {
	Name            string        `env:"REDIS_NAME" yaml:"name"`
	Addr            string        `env:"REDIS_ADDR" yaml:"addr"`
	Password        string        `env:"REDIS_PASSWORD" yaml:"password"`
	DB              int           `env:"REDIS_DB" yaml:"db"`
	PoolSize        int           `env:"REDIS_POOL_SIZE" yaml:"pool_size"`
	PoolTimeout     time.Duration `env:"REDIS_POOL_TIMEOUT" yaml:"pool_timeout"`
	ReadTimeout     time.Duration `env:"REDIS_READ_TIMEOUT" yaml:"read_timeout"`
	WriteTimeout    time.Duration `env:"REDIS_WRITE_TIMEOUT" yaml:"write_timeout"`
	IdleTimeout     time.Duration `env:"REDIS_IDLE_TIMEOUT" yaml:"idle_timeout"`
	MaxRetries      int           `env:"REDIS_MAX_RETRIES" yaml:"max_retries"`
	MinRetryBackoff time.Duration `env:"REDIS_MIN_RETRY_BACKOFF" yaml:"min_retry_backoff"`
	MaxRetryBackoff time.Duration `env:"REDIS_MAX_RETRY_BACKOFF" yaml:"max_retry_backoff"`
	DialTimeout     time.Duration `env:"REDIS_DIAL_TIMEOUT" yaml:"dial_timeout"`
}

var client *redis.Client

func Init(opt *redis.Options) error {
	client = redis.NewClient(opt)
	return client.Ping(context.Background()).Err()
}

var clients sync.Map

func Inits(conf []*Options) error {
	for _, conf := range conf {
		if conf.Name == "" {
			conf.Name = "default"
		}
		opt := &redis.Options{
			Addr:         conf.Addr,
			Password:     conf.Password,
			DB:           conf.DB,
			PoolSize:     conf.PoolSize,
			PoolTimeout:  conf.PoolTimeout,
			ReadTimeout:  conf.ReadTimeout,
			WriteTimeout: conf.WriteTimeout,
			IdleTimeout:  conf.IdleTimeout,
		}
		if err := Init(opt); err != nil {
			return err
		}
		clients.Store(conf.Name, client)
	}
	return nil
}

func Get() *redis.Client {
	return GetClient("default")
}

func GetClient(name string) *redis.Client {
	if client, ok := clients.Load(name); ok {
		return client.(*redis.Client)
	}
	return nil
}
