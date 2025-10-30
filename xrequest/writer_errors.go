package xrequest

import (
	"context"
	"errors"
	"net"
	"strings"
	"syscall"
)

// IsClientDisconnected 判断写入响应时的错误是否由下游连接中断引起。
// 常见的场景包括：浏览器主动取消、客户端超时、反向代理提前断开连接等。
func IsClientDisconnected(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	if errors.Is(err, net.ErrClosed) {
		return true
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		if opErr.Err != nil {
			if errors.Is(opErr.Err, syscall.EPIPE) || errors.Is(opErr.Err, syscall.ECONNRESET) {
				return true
			}
		}
	}

	if errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
		return true
	}

	errMsg := err.Error()
	if strings.Contains(errMsg, "broken pipe") || strings.Contains(errMsg, "connection reset by peer") {
		return true
	}

	return false
}
