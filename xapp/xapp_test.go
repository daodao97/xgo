package xapp

import (
	"errors"
	"os"
	"sync"
	"testing"
	"time"
)

type readyTestServer struct {
	started    chan struct{}
	run        chan struct{}
	stopOnce   sync.Once
	stopCalled chan struct{}
}

func newReadyTestServer() *readyTestServer {
	return &readyTestServer{
		started:    make(chan struct{}),
		run:        make(chan struct{}),
		stopCalled: make(chan struct{}),
	}
}

func (s *readyTestServer) Start() error {
	<-s.run
	return nil
}

func (s *readyTestServer) Stop() {
	s.stopOnce.Do(func() {
		close(s.run)
		close(s.stopCalled)
	})
}

func (s *readyTestServer) Started() <-chan struct{} {
	return s.started
}

type errorReadyServer struct{}

func (errorReadyServer) Start() error {
	return errors.New("boom")
}

func (errorReadyServer) Stop() {}

func (errorReadyServer) Started() <-chan struct{} {
	return make(chan struct{})
}

func TestAppRunWaitsForReadyServersBeforeAfterStarted(t *testing.T) {
	s1 := newReadyTestServer()
	s2 := newReadyTestServer()
	afterStarted := make(chan struct{})
	runDone := make(chan error, 1)
	injectedSignals := make(chan os.Signal, 1)

	oldNotify := signalNotify
	oldStop := signalStop
	signalNotify = func(c chan<- os.Signal, _ ...os.Signal) {
		go func() {
			sig := <-injectedSignals
			c <- sig
		}()
	}
	signalStop = func(chan<- os.Signal) {}
	defer func() {
		signalNotify = oldNotify
		signalStop = oldStop
	}()

	app := NewApp().
		AddServer(func() Server { return s1 }, func() Server { return s2 }).
		AfterStarted(func() {
			close(afterStarted)
		})

	go func() {
		runDone <- app.Run()
	}()

	select {
	case <-afterStarted:
		t.Fatal("afterStarted called before servers became ready")
	case <-time.After(50 * time.Millisecond):
	}

	close(s1.started)
	close(s2.started)

	select {
	case <-afterStarted:
	case <-time.After(time.Second):
		t.Fatal("afterStarted was not called after readiness signals")
	}

	injectedSignals <- os.Interrupt

	select {
	case err := <-runDone:
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run did not exit after shutdown signal")
	}

	select {
	case <-s1.stopCalled:
	case <-time.After(time.Second):
		t.Fatal("server 1 was not stopped")
	}

	select {
	case <-s2.stopCalled:
	case <-time.After(time.Second):
		t.Fatal("server 2 was not stopped")
	}
}

func TestAppRunSkipsAfterStartedWhenReadyServerFails(t *testing.T) {
	called := false

	err := NewApp().
		AddServer(func() Server { return errorReadyServer{} }).
		AfterStarted(func() {
			called = true
		}).
		Run()

	if err == nil {
		t.Fatal("expected startup error")
	}
	if err.Error() != "server error: boom" {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("afterStarted should not run when startup fails")
	}
}
