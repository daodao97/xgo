package xutil

import (
	"context"

	"github.com/daodao97/xgo/xlog"
)

func Go(ctx context.Context, fn func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				xlog.ErrorCtx(ctx, "safe go", xlog.Any("error", err))
			}
		}()
		fn()
	}()
}
