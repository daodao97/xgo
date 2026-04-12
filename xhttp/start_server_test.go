package xhttp

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestShutdownServerUsesTimeoutContext(t *testing.T) {
	timeout := 50 * time.Millisecond
	var deadline time.Time

	shutdownServer(timeout, func(ctx context.Context) error {
		var ok bool
		deadline, ok = ctx.Deadline()
		if !ok {
			t.Fatal("expected shutdown context to have deadline")
		}
		return nil
	}, func(int) {
		t.Fatal("did not expect exit on successful shutdown")
	})

	remaining := time.Until(deadline)
	if remaining <= 0 || remaining > timeout {
		t.Fatalf("unexpected shutdown deadline window: %v", remaining)
	}
}

func TestShutdownServerExitsOnTimeout(t *testing.T) {
	exitCode := 0

	shutdownServer(time.Millisecond, func(ctx context.Context) error {
		return context.DeadlineExceeded
	}, func(code int) {
		exitCode = code
	})

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
}

func TestShutdownServerDoesNotExitOnNonTimeoutError(t *testing.T) {
	exitCode := 0

	shutdownServer(time.Millisecond, func(ctx context.Context) error {
		return errors.New("shutdown failed")
	}, func(code int) {
		exitCode = code
	})

	if exitCode != 0 {
		t.Fatalf("expected no exit, got code %d", exitCode)
	}
}
