package xapp

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/daodao97/xgo/xlog"
)

type NewServer func() Server

func NewHttp(addr string, handler func() http.Handler) NewServer {
	return func() Server {
		return NewHTTPServer(addr, handler)
	}
}

type HTTPServer struct {
	server *http.Server
}

func NewHTTPServer(addr string, handler func() http.Handler) *HTTPServer {
	if handler == nil {
		handler = func() http.Handler {
			return http.DefaultServeMux
		}
	}

	return &HTTPServer{
		server: &http.Server{
			Addr:    addr,
			Handler: handler(),
		},
	}
}

func (s *HTTPServer) Start() error {
	xlog.Debug("Starting HTTP server on", xlog.String("port", s.server.Addr))
	return s.server.ListenAndServe()
}

func (s *HTTPServer) Stop() {
	xlog.Debug("Stopping HTTP server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		xlog.Error("HTTP server Shutdown", xlog.Err(err))
	} else {
		xlog.Debug("HTTP server Shutdown completed successfully")
	}

	// 确认端口是否关闭
	time.Sleep(1 * time.Second) // 等待1秒以确保所有连接已关闭
	conn, err := net.DialTimeout("tcp", s.server.Addr, 1*time.Second)
	if err == nil {
		conn.Close()
		xlog.Error("HTTP server is still accessible on", xlog.String("port", s.server.Addr))
	} else {
		xlog.Debug("HTTP server is not accessible anymore")
	}

	xlog.Debug("Stop HTTP server done")
}
