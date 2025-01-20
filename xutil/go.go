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

func GoWithCancel(ctx context.Context, fn func(c context.Context)) context.CancelFunc {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		defer cancel()
		fn(ctx)
	}()
	return cancel
}
