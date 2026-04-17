package xapp

import (
	"fmt"
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
