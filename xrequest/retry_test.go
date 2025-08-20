package xrequest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetryFunctionality(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt32(&attempts, 1)
		if current < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Attempt %d failed", current)
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Success on attempt %d", current)
		}
	}))
	defer server.Close()

	resp, err := New().
		SetRetry(3, 100*time.Millisecond).
		Get(server.URL)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if resp.StatusCode() != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode())
	}

	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("Expected 3 attempts, got %d", atomic.LoadInt32(&attempts))
	}
}

func TestRetryWithNetworkError(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Success")
	}))
	// Close server immediately to simulate network error initially
	server.Close()

	_, err := New().
		SetRetry(2, 50*time.Millisecond).
		Get(server.URL)

	if err == nil {
		t.Fatal("Expected error due to closed server")
	}

	// Should have attempted 2 times
	if atomic.LoadInt32(&attempts) != 0 {
		t.Errorf("Expected 0 successful attempts, got %d", atomic.LoadInt32(&attempts))
	}
}

func TestRetryWithCustomCondition(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt32(&attempts, 1)
		if current < 2 {
			w.WriteHeader(http.StatusBadRequest) // 400 - normally wouldn't retry
			fmt.Fprintf(w, "Bad request attempt %d", current)
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Success on attempt %d", current)
		}
	}))
	defer server.Close()

	// Custom retry condition: retry on 400 status
	resp, err := New().
		SetRetryWithCondition(3, 50*time.Millisecond, func(resp *Response, err error) bool {
			return resp != nil && resp.StatusCode() == 400
		}).
		Get(server.URL)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if resp.StatusCode() != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode())
	}

	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("Expected 2 attempts, got %d", atomic.LoadInt32(&attempts))
	}
}

func TestRetryExhaustion(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Always fails")
	}))
	defer server.Close()

	resp, err := New().
		SetRetry(2, 50*time.Millisecond).
		Get(server.URL)

	// Should return the response and descriptive error
	if err == nil {
		t.Fatal("Expected error after retry exhaustion")
	}

	if resp == nil {
		t.Fatal("Expected response even after failure")
	}

	if resp.StatusCode() != 500 {
		t.Errorf("Expected status 500, got %d", resp.StatusCode())
	}

	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("Expected 2 attempts, got %d", atomic.LoadInt32(&attempts))
	}

	// Check error message contains attempt count and status
	expectedErrMsg := "request failed after 2 attempts, status: 500"
	if err.Error() != expectedErrMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, err.Error())
	}
}