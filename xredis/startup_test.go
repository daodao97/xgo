package xredis

import (
	"strings"
	"testing"
	"time"
)

func TestOptionsRedisOptionsFromDSN(t *testing.T) {
	conf := Options{
		DSN:      "redis://user:secret@redis.example:6380/3?pool_size=12&read_timeout=2s",
		Addr:     "ignored:6379",
		Password: "ignored",
		DB:       9,
	}

	opt, err := conf.RedisOptions()
	if err != nil {
		t.Fatalf("RedisOptions failed: %v", err)
	}
	if opt.Addr != "redis.example:6380" {
		t.Fatalf("expected DSN address, got %q", opt.Addr)
	}
	if opt.Username != "user" || opt.Password != "secret" {
		t.Fatalf("expected DSN credentials, got username %q and password %q", opt.Username, opt.Password)
	}
	if opt.DB != 3 {
		t.Fatalf("expected DSN database 3, got %d", opt.DB)
	}
	if opt.PoolSize != 12 {
		t.Fatalf("expected DSN pool size 12, got %d", opt.PoolSize)
	}
	if opt.ReadTimeout != 2*time.Second {
		t.Fatalf("expected DSN read timeout 2s, got %s", opt.ReadTimeout)
	}
}

func TestOptionsRedisOptionsFromFields(t *testing.T) {
	conf := Options{
		Addr:        "redis:6379",
		Password:    "secret",
		DB:          2,
		PoolSize:    10,
		ReadTimeout: time.Second,
	}

	opt, err := conf.RedisOptions()
	if err != nil {
		t.Fatalf("RedisOptions failed: %v", err)
	}
	if opt.Addr != conf.Addr || opt.Password != conf.Password || opt.DB != conf.DB {
		t.Fatalf("expected field-based connection options, got %+v", opt)
	}
	if opt.PoolSize != conf.PoolSize || opt.ReadTimeout != conf.ReadTimeout {
		t.Fatalf("expected field-based pool options, got %+v", opt)
	}
}

func TestOptionsRedisOptionsRejectsInvalidDSN(t *testing.T) {
	_, err := (Options{DSN: "mysql://localhost/db"}).RedisOptions()
	if err == nil || !strings.Contains(err.Error(), "invalid URL scheme") {
		t.Fatalf("expected invalid DSN scheme error, got %v", err)
	}
}

func TestOptionsRedisOptionsRejectsDSNInClusterMode(t *testing.T) {
	_, err := (Options{DSN: "redis://redis:6379", IsCluster: true}).RedisOptions()
	if err == nil || !strings.Contains(err.Error(), "cluster mode") {
		t.Fatalf("expected cluster mode error, got %v", err)
	}
}
