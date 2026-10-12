package xapp

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestNewGinHttpServerStartDoesNotPanic(t *testing.T) {
	server, ok := NewGinHttpServer("127.0.0.1:0", func() *gin.Engine {
		return gin.New()
	})().(*HTTPServer)
	if !ok {
		t.Fatal("expected *HTTPServer")
	}

	done := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("panic: %v", r)
			}
		}()
		done <- server.Start()
	}()

	select {
	case <-server.Started():
	case <-time.After(time.Second):
		t.Fatal("server did not report started")
	}

	server.Stop()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("server exited with error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("server did not stop")
	}
}

func TestHTTPShutdownTimeout(t *testing.T) {
	for _, tc := range []struct {
		name    string
		timeout time.Duration
		want    time.Duration
	}{
		{"default", 0, 30 * time.Second},
		{"negative keeps default", -time.Second, 30 * time.Second},
		{"configured", 5 * time.Second, 5 * time.Second},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, server := range []*HTTPServer{
				NewHTTPServer(":0", nil, WithHTTPShutdownTimeout(tc.timeout)),
				NewHttp(":0", nil, WithHTTPShutdownTimeout(tc.timeout))().(*HTTPServer),
				NewGinHttpServer(":0", func() *gin.Engine { return gin.New() }, WithHTTPShutdownTimeout(tc.timeout))().(*HTTPServer),
			} {
				assert.Equal(t, tc.want, server.shutdownTimeout)
			}
		})
	}
}

func TestHTTPStop_TimeoutClosesActiveRequest(t *testing.T) {
	entered := make(chan struct{})
	canceled := make(chan struct{})
	server := NewHTTPServer("", func() http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			close(entered)
			<-r.Context().Done()
			close(canceled)
		})
	}, WithHTTPShutdownTimeout(10*time.Millisecond))
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = server.server.Close() })
	serveDone := make(chan error, 1)
	go func() { serveDone <- server.server.Serve(ln) }()
	requestDone := make(chan struct{})
	go func() {
		defer close(requestDone)
		resp, _ := http.Get("http://" + ln.Addr().String())
		if resp != nil {
			_ = resp.Body.Close()
		}
	}()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("request did not enter handler")
	}
	stopped := make(chan struct{})
	go func() { server.Stop(); close(stopped) }()
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("Stop did not use the configured deadline")
	}
	select {
	case <-canceled:
	case <-time.After(time.Second):
		t.Fatal("active connection was not closed")
	}
	select {
	case err := <-serveDone:
		require.ErrorIs(t, err, http.ErrServerClosed)
	case <-time.After(time.Second):
		t.Fatal("Serve did not stop")
	}
	select {
	case <-requestDone:
	case <-time.After(time.Second):
		t.Fatal("client request did not stop")
	}
}
