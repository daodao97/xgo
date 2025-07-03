package xutil

import (
	"context"
	"fmt"
	"time"
)

type RetryOption struct {
	MaxRetries int
	Delay      time.Duration
}

type RetryOptionFunc func(opt *RetryOption)

func WithMaxRetries(maxRetries int) RetryOptionFunc {
	return func(opt *RetryOption) {
		opt.MaxRetries = maxRetries
	}
}

func WithDelay(delay time.Duration) RetryOptionFunc {
	return func(opt *RetryOption) {
		opt.Delay = delay
	}
}

func Retry[T any](ctx context.Context, fn func(ctx context.Context) (T, error), opts ...RetryOptionFunc) (T, error) {
	var result T
	var err error

	opt := RetryOption{
		MaxRetries: 3,
		Delay:      1 * time.Second,
	}
	for _, optFn := range opts {
		optFn(&opt)
	}
	for i := 0; i < opt.MaxRetries; i++ {
		result, err = fn(ctx)
		if err == nil {
			return result, nil
		}
		time.Sleep(opt.Delay)
	}
	return result, fmt.Errorf("retry failed: %w", err)
}
