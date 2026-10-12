package xapp

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/daodao97/xgo/xlog"
)

type NewServer func() Server

func NewHttp(addr string, handler func() http.Handler, opts ...HTTPServerOption) NewServer {
	return func() Server {
		return NewHTTPServer(addr, handler, opts...)
	}
}

// HTTPServerOption configures an HTTP server before it starts.
type HTTPServerOption func(*HTTPServer)

// WithHTTPShutdownTimeout sets the graceful shutdown budget. Non-positive values
// retain the default of 30 seconds.
func WithHTTPShutdownTimeout(timeout time.Duration) HTTPServerOption {
	return func(s *HTTPServer) {
		if timeout > 0 {
			s.shutdownTimeout = timeout
		}
	}
}

type HTTPServer struct {
	shutdownTimeout time.Duration
	server          *http.Server
	started         chan struct{}
	startedOnce     sync.Once
}

func NewHTTPServer(addr string, handler func() http.Handler, opts ...HTTPServerOption) *HTTPServer {
	if handler == nil {
		handler = func() http.Handler {
			return http.DefaultServeMux
		}
	}

	s := &HTTPServer{
		shutdownTimeout: 30 * time.Second,
		server: &http.Server{
			Addr:    addr,
			Handler: handler(),
		},
		started: make(chan struct{}),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *HTTPServer) Start() error {
	xlog.Debug("Starting HTTP server on", xlog.String("port", s.server.Addr))
	ln, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return err
	}
	s.startedOnce.Do(func() {
		close(s.started)
	})
	err = s.server.Serve(ln)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *HTTPServer) Started() <-chan struct{} {
	return s.started
}

func (s *HTTPServer) Stop() {
	xlog.Debug("Stopping HTTP server")
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		if err == context.DeadlineExceeded {
			xlog.Warn("HTTP server graceful shutdown timed out, forcing close",
				xlog.Duration("timeout", s.shutdownTimeout))
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
