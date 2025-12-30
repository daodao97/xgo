package xrequest

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestToHttpResponseWriteV2_SSESkipEmptyLineHook(t *testing.T) {
	body := "data: a\n\n"
	rawResp := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"text/event-stream"},
		},
		Body: io.NopCloser(strings.NewReader(body)),
	}
	resp := NewResponse(rawResp)

	var seen [][]byte
	hook := func(data []byte) (bool, []byte) {
		seen = append(seen, append([]byte(nil), data...))
		return true, data
	}

	rec := httptest.NewRecorder()
	written, err := resp.ToHttpResponseWriteV2(rec, hook)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got, want := len(seen), 1; got != want {
		t.Fatalf("hook calls mismatch: got %d want %d", got, want)
	}
	if !bytes.Equal(seen[0], []byte("data: a")) {
		t.Fatalf("first hook data mismatch: got %q", string(seen[0]))
	}
	if got := rec.Body.String(); got != body {
		t.Fatalf("body mismatch: got %q want %q", got, body)
	}
	if written != int64(len(body)) {
		t.Fatalf("written mismatch: got %d want %d", written, len(body))
	}
}

func TestToHttpResponseWriteV2_NonStreamContentLength(t *testing.T) {
	rawResp := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			"Content-Length": []string{"3"},
		},
		Body: io.NopCloser(strings.NewReader("abc")),
	}
	resp := NewResponse(rawResp)

	hook := func(data []byte) (bool, []byte) {
		return true, append(data, 'd')
	}

	rec := httptest.NewRecorder()
	_, err := resp.ToHttpResponseWriteV2(rec, hook)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := rec.Body.String(); got != "abcd" {
		t.Fatalf("body mismatch: got %q want %q", got, "abcd")
	}
	if got := rec.Header().Get("Content-Length"); got != "4" {
		t.Fatalf("content-length mismatch: got %q want %q", got, "4")
	}
}

func TestToHttpResponseWriteV2_NonStreamFlushFalse(t *testing.T) {
	rawResp := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(strings.NewReader("abc")),
	}
	resp := NewResponse(rawResp)

	hook := func(data []byte) (bool, []byte) {
		return false, data
	}

	rec := httptest.NewRecorder()
	_, err := resp.ToHttpResponseWriteV2(rec, hook)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := rec.Body.Len(); got != 0 {
		t.Fatalf("body length mismatch: got %d want %d", got, 0)
	}
	if got := rec.Header().Get("Content-Length"); got != "0" {
		t.Fatalf("content-length mismatch: got %q want %q", got, "0")
	}
}
