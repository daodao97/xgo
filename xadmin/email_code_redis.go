package xadmin

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const defaultLoginEmailCodeRedisPrefix = "xadmin:login_email_code"

var loginEmailCodeRedisConsumeScript = redis.NewScript(`
local code = redis.call("GET", KEYS[1])
if not code then
	return 0
end
if code == ARGV[1] then
	redis.call("DEL", KEYS[1])
	redis.call("DEL", KEYS[2])
	return 1
end
return 0
`)

type LoginEmailCodeRedisStoreOption func(*LoginEmailCodeRedisStore)

func WithLoginEmailCodeRedisPrefix(prefix string) LoginEmailCodeRedisStoreOption {
	return func(s *LoginEmailCodeRedisStore) {
		if strings.TrimSpace(prefix) != "" {
			s.prefix = strings.TrimRight(strings.TrimSpace(prefix), ":")
		}
	}
}

type LoginEmailCodeRedisStore struct {
	client redis.UniversalClient
	prefix string
}

func NewLoginEmailCodeRedisStore(client redis.UniversalClient, opts ...LoginEmailCodeRedisStoreOption) *LoginEmailCodeRedisStore {
	store := &LoginEmailCodeRedisStore{
		client: client,
		prefix: defaultLoginEmailCodeRedisPrefix,
	}
	for _, opt := range opts {
		opt(store)
	}
	return store
}

func SetLoginEmailCodeRedisStore(client redis.UniversalClient, opts ...LoginEmailCodeRedisStoreOption) {
	SetLoginEmailCodeStore(NewLoginEmailCodeRedisStore(client, opts...))
}

func (s *LoginEmailCodeRedisStore) codeKey(email string) string {
	return s.prefix + ":{" + email + "}:code"
}

func (s *LoginEmailCodeRedisStore) cooldownKey(email string) string {
	return s.prefix + ":{" + email + "}:cooldown"
}

func (s *LoginEmailCodeRedisStore) checkClient() error {
	if s == nil || s.client == nil {
		return errors.New("redis client not initialized")
	}
	return nil
}

func (s *LoginEmailCodeRedisStore) RemainingCooldown(ctx context.Context, email string, interval time.Duration, now time.Time) (time.Duration, error) {
	if interval <= 0 {
		return 0, nil
	}
	if err := s.checkClient(); err != nil {
		return 0, err
	}

	ttl, err := s.client.TTL(ctx, s.cooldownKey(email)).Result()
	if err != nil {
		return 0, err
	}
	if ttl <= 0 {
		return 0, nil
	}
	return ttl, nil
}

func (s *LoginEmailCodeRedisStore) Save(ctx context.Context, email, code string, codeExpire, sendInterval time.Duration, now time.Time) error {
	if err := s.checkClient(); err != nil {
		return err
	}

	_, err := s.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, s.codeKey(email), code, codeExpire)
		if sendInterval > 0 {
			pipe.Set(ctx, s.cooldownKey(email), "1", sendInterval)
		}
		return nil
	})
	return err
}

func (s *LoginEmailCodeRedisStore) Consume(ctx context.Context, email, code string, now time.Time) (bool, error) {
	if err := s.checkClient(); err != nil {
		return false, err
	}

	result, err := loginEmailCodeRedisConsumeScript.Run(ctx, s.client, []string{
		s.codeKey(email),
		s.cooldownKey(email),
	}, code).Int()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}
