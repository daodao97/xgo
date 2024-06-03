package xhttp

import (
	"context"
	"errors"
	"github.com/daodao97/xgo/xlog"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func StartServer(handler http.Handler, addr string, shutdownTimeout time.Duration) {
	// 创建一个自定义的 HTTP 服务器
	srv := &http.Server{
		Handler: handler,
		Addr:    addr,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
		//IdleTimeout:  30 * time.Second,
	}

	// 启动服务器
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			xlog.Error("listen: " + err.Error())
		}
	}()

	xlog.Info("Server started on " + addr)

	// 创建一个通道，用于接收操作系统信号
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt)

	// 阻塞，直到收到中断信号
	<-stopChan
	xlog.Info("Shutting down xhttp...")

	// 创建一个上下文，设置关闭服务器的超时时间
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// 优雅地关闭服务器
	if err := srv.Shutdown(ctx); err != nil {
		xlog.Error("Server Shutdown Failed: " + err.Error())
	}
	xlog.Info("Server gracefully stopped")
}
