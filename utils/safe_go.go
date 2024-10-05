package utils

import "github.com/daodao97/xgo/xlog"

func SafeGo(f func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				xlog.Error("safe go panic", xlog.Any("error", err))
			}
		}()
		f()
	}()
}
