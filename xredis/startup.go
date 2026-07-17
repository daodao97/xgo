package xredis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type Options struct {
	Name            string        `env:"NAME" yaml:"name"`
	DSN             string        `env:"DSN" yaml:"dsn"`
	Addr            string        `env:"ADDR" yaml:"addr"`
	Password        string        `env:"PASSWORD" yaml:"password"`
	DB              int           `env:"DB" yaml:"db"`
	PoolSize        int           `env:"POOL_SIZE" yaml:"pool_size"`
	PoolTimeout     time.Duration `env:"POOL_TIMEOUT" yaml:"pool_timeout"`
	ReadTimeout     time.Duration `env:"READ_TIMEOUT" yaml:"read_timeout"`
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT" yaml:"write_timeout"`
	IdleTimeout     time.Duration `env:"IDLE_TIMEOUT" yaml:"idle_timeout"`
	MaxRetries      int           `env:"MAX_RETRIES" yaml:"max_retries"`
	MinRetryBackoff time.Duration `env:"MIN_RETRY_BACKOFF" yaml:"min_retry_backoff"`
	MaxRetryBackoff time.Duration `env:"MAX_RETRY_BACKOFF" yaml:"max_retry_backoff"`
	DialTimeout     time.Duration `env:"DIAL_TIMEOUT" yaml:"dial_timeout"`

	// Cluster mode settings
	IsCluster       bool     `env:"IS_CLUSTER" yaml:"is_cluster"`
	ClusterAddrs    []string `env:"CLUSTER_ADDRS" yaml:"cluster_addrs"`
	ClusterPassword string   `env:"CLUSTER_PASSWORD" yaml:"cluster_password"`
}

// RedisOptions converts the configuration into go-redis options.
// When DSN is set, it is treated as the complete single-node configuration and
// the individual connection fields are ignored. Supported schemes are redis,
// rediss, and unix; query parameters are parsed by go-redis.
func (o Options) RedisOptions() (*redis.Options, error) {
	if o.DSN != "" {
		if o.IsCluster {
			return nil, fmt.Errorf("redis DSN is not supported in cluster mode")
		}

		opt, err := redis.ParseURL(o.DSN)
		if err != nil {
			return nil, fmt.Errorf("parse redis DSN: %w", err)
		}
		return opt, nil
	}

	return &redis.Options{
		Addr:            o.Addr,
		Password:        o.Password,
		DB:              o.DB,
		PoolSize:        o.PoolSize,
		PoolTimeout:     o.PoolTimeout,
		ReadTimeout:     o.ReadTimeout,
		WriteTimeout:    o.WriteTimeout,
		ConnMaxIdleTime: o.IdleTimeout,
		MaxRetries:      o.MaxRetries,
		MinRetryBackoff: o.MinRetryBackoff,
		MaxRetryBackoff: o.MaxRetryBackoff,
		DialTimeout:     o.DialTimeout,
	}, nil
}

var client redis.UniversalClient

func Init(opt *redis.Options) error {
	client = redis.NewClient(opt)
	return client.Ping(context.Background()).Err()
}

func InitCluster(opt *redis.ClusterOptions) error {
	client = redis.NewClusterClient(opt)
	return client.Ping(context.Background()).Err()
}

func InitUniversal(opt *redis.UniversalOptions) error {
	client = redis.NewUniversalClient(opt)
	return client.Ping(context.Background()).Err()
}

var clients sync.Map

func Inits(conf []Options) error {
	for _, conf := range conf {
		if conf.Name == "" {
			conf.Name = "default"
		}

		var c redis.UniversalClient
		var err error

		if conf.IsCluster && conf.DSN != "" {
			return fmt.Errorf("redis DSN is not supported in cluster mode")
		}

		if conf.IsCluster {
			addrs := conf.ClusterAddrs
			if len(addrs) == 0 && conf.Addr != "" {
				addrs = []string{conf.Addr}
			}
			clusterOpt := &redis.ClusterOptions{
				Addrs:           addrs,
				Password:        conf.ClusterPassword,
				PoolSize:        conf.PoolSize,
				PoolTimeout:     conf.PoolTimeout,
				ReadTimeout:     conf.ReadTimeout,
				WriteTimeout:    conf.WriteTimeout,
				ConnMaxIdleTime: conf.IdleTimeout,
				MaxRetries:      conf.MaxRetries,
				MinRetryBackoff: conf.MinRetryBackoff,
				MaxRetryBackoff: conf.MaxRetryBackoff,
				DialTimeout:     conf.DialTimeout,
			}
			c = redis.NewClusterClient(clusterOpt)
		} else {
			opt, parseErr := conf.RedisOptions()
			if parseErr != nil {
				return parseErr
			}
			c = redis.NewClient(opt)
		}

		if err = c.Ping(context.Background()).Err(); err != nil {
			return err
		}

		if conf.Name == "default" {
			client = c
		}
		clients.Store(conf.Name, c)
	}
	return nil
}

func Get() redis.UniversalClient {
	return client
}

func GetClient(name string) redis.UniversalClient {
	if c, ok := clients.Load(name); ok {
		return c.(redis.UniversalClient)
	}
	return nil
}
