package xutil

import (
	"context"
	"runtime/debug"

	"github.com/daodao97/xgo/xlog"
)

func Go(ctx context.Context, fn func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()
				xlog.ErrorCtx(ctx, "safe go",
					xlog.Any("error", err),
					xlog.String("stack", string(stack)),
				)
			}
		}()
		fn()
	}()
}

// GoCtx is same as Go but with context
//
// The context passed to fn is without cancel,
// so that it won't be canceled when the parent context is canceled.
// It is useful for web server to avoid canceling the background task when the request is canceled.
func GoCtx(ctx context.Context, fn func(c context.Context)) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()
				xlog.ErrorCtx(ctx, "safe go",
					xlog.Any("error", err),
					xlog.String("stack", string(stack)),
				)
			}
		}()
		ctx = context.WithoutCancel(ctx)
		fn(ctx)
	}()
}

func GoWithCancel(ctx context.Context, fn func(c context.Context)) context.CancelFunc {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		defer cancel()
		fn(ctx)
	}()
	return cancel
}
