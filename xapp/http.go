package xapp

import (
	"context"
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		if err == context.DeadlineExceeded {
			xlog.Warn("HTTP server graceful shutdown timed out, forcing close",
				xlog.Duration("timeout", 30*time.Second))
			if err := s.server.Close(); err != nil {
				xlog.Error("HTTP server force close failed", xlog.Err(err))
			}
		} else {
			xlog.Error("HTTP server shutdown error", xlog.Err(err))
		}
	} else {
		xlog.Debug("HTTP server shutdown completed successfully")
	}

	xlog.Debug("Stop HTTP server done")
}
