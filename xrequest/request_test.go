package xrequest

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/daodao97/xgo/xjson"
)

type flushRecorder struct {
	*httptest.ResponseRecorder
	t *testing.T
}

func newFlushRecorder(t *testing.T) *flushRecorder {
	return &flushRecorder{ResponseRecorder: httptest.NewRecorder(), t: t}
}

func (fr *flushRecorder) Flush() {}

func (fr *flushRecorder) Write(b []byte) (int, error) {
	if fr.t != nil {
		fr.t.Logf("flushRecorder chunk: %s", bytes.TrimRight(b, "\n"))
	} else {
		fmt.Printf("flushRecorder chunk: %s\n", string(bytes.TrimRight(b, "\n")))
	}
	return fr.ResponseRecorder.Write(b)
}

func TestRequest(t *testing.T) {
	b := `{
		"name": "daodao"
	}`

	request := New().
		SetMethod("POST").
		SetHeaders(map[string]string{
			"Content-Type": "application/json",
			"X-TEST":       "test",
			"x-app":        "test",
			"accept":       "x-test",
		}).
		SetCookies(map[string]string{
			"test":  "test",
			"test2": "test2",
		}).
		SetBody(b).
		SetURL("https://httpbin.org/post").
		SetRetry(3, time.Second)
	resp, err := request.Do()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(resp.StatusCode())
	t.Log(resp.Json())
}

func TestRequestWithQueryParams(t *testing.T) {

	queryParams := struct {
		Name string `json:"name"`
		Age  int    `json:"age,omitempty"`
	}{
		Name: "daodao",
		Age:  0,
	}

	req := New().
		SetMethod(http.MethodGet).
		SetURL("https://httpbin.org/get").
		SetQueryParams(xjson.New(queryParams).MapString())
	resp, err := req.Do()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp.StatusCode())
	t.Log(resp.Json())
}

func TestRequestWithFile(t *testing.T) {
	file, err := os.Open("/Users/daodao/Downloads/sample.pdf")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	req := New().
		SetMethod(http.MethodPost).
		SetURL("http://127.0.0.1:8000").
		AddFile("file", "sample.pdf", file)

	resp, err := req.Do()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp.StatusCode())
	t.Log(resp.Json())
}

func TestRequestWithReqHook(t *testing.T) {
	req := New().
		SetMethod(http.MethodPost).
		SetURL("http://127.0.0.1:8000").
		AddReqHook(func(req *http.Request) error {
			req.Header.Add("X-Test", "test")
			return nil
		})
	resp, err := req.Do()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp.StatusCode())
	t.Log(resp.Json())
}

func TestRequestSSE(t *testing.T) {
	// 启动测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		// 发送10条测试消息
		for i := 0; i < 10; i++ {
			fmt.Fprintf(w, "data: Message %d\n\n", i)
			flusher.Flush()
			time.Sleep(1 * time.Second)
		}
	}))
	defer server.Close()

	req := New().
		SetMethod(http.MethodGet).
		SetURL(server.URL) // 使用测试服务器的URL
	resp, err := req.Do()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp.StatusCode())

	ch, err := resp.Stream()
	if err != nil {
		t.Fatal(err)
	}

	// 收集所有消息
	var messages []string
	for msg := range ch {
		messages = append(messages, msg)
		t.Log(msg)
	}

	// 验证接收到的消息
	if len(messages) != 10 {
		t.Errorf("Expected 10 messages, got %d", len(messages))
	}
	for i, msg := range messages {
		expected := fmt.Sprintf("Message %d", i)
		if msg != expected {
			t.Errorf("Expected message %q, got %q", expected, msg)
		}
	}
}

func TestRequestWithDebug(t *testing.T) {
	req := New().SetDebug(true)
	resp, err := req.SetRetry(3, time.Second).Post("http://127.0.0.1:8000/callback")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Error() != nil {
		t.Fatal(resp.Error())
	}
}

func TestRequestWithProxy(t *testing.T) {
	// export HTTP_PROXY=http://127.0.0.1:8000
	fmt.Println("HTTP_PROXY", os.Getenv("HTTP_PROXY"))
	fmt.Println("HTTPS_PROXY", os.Getenv("HTTPS_PROXY"))
	fmt.Println("ALL_PROXY", os.Getenv("ALL_PROXY"))

	req := New()
	resp, err := req.SetRetry(3, time.Second).Post("https://ipinfo.io")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp.StatusCode())
	t.Log(resp.Json())
}

func TestRequestWithProxyWarp(t *testing.T) {
	// export HTTP_PROXY=http://127.0.0.1:8000
	fmt.Println("HTTP_PROXY", os.Getenv("HTTP_PROXY"))
	fmt.Println("HTTPS_PROXY", os.Getenv("HTTPS_PROXY"))
	fmt.Println("ALL_PROXY", os.Getenv("ALL_PROXY"))

	req := New()
	resp, err := req.SetRetry(3, time.Second).Post("https://cloudflare.com/cdn-cgi/trace")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp.StatusCode())
	t.Log(resp.Json())
}

func TestResponseToHttpResponseWriter(t *testing.T) {
	const responseBody = `{"message":"ok"}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Test-Header", "test-value")
		fmt.Fprint(w, responseBody)
	}))
	defer server.Close()

	resp, err := New().Get(server.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	recorder := httptest.NewRecorder()
	bodyLen, err := resp.ToHttpResponseWriter(recorder)
	if err != nil {
		t.Fatalf("write response failed: %v", err)
	}

	if bodyLen != int64(len(responseBody)) {
		t.Fatalf("unexpected body length, expected %d got %d", len(responseBody), bodyLen)
	}

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status code, expected 200 got %d", recorder.Code)
	}

	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("unexpected content-type header, got %s", got)
	}

	if got := recorder.Header().Get("X-Test-Header"); got != "test-value" {
		t.Fatalf("unexpected custom header, got %s", got)
	}

	if got := recorder.Body.String(); got != responseBody {
		t.Fatalf("unexpected body, got %q", got)
	}
}

func TestResponseToHttpResponseWriterStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatalf("server does not support flushing")
		}

		for i := 0; i < 3; i++ {
			fmt.Fprintf(w, "data: message-%d\n\n", i)
			flusher.Flush()
		}
	}))
	defer server.Close()

	resp, err := New().Get(server.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	recorder := newFlushRecorder(t)

	bodyLen, err := resp.ToHttpResponseWriter(recorder)
	if err != nil {
		t.Fatalf("stream write failed: %v", err)
	}

	const expected = "data: message-0\n\ndata: message-1\n\ndata: message-2\n\n"
	if bodyLen != int64(len(expected)) {
		t.Fatalf("unexpected body length, expected %d got %d", len(expected), bodyLen)
	}

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status code, expected 200 got %d", recorder.Code)
	}

	if got := recorder.Header().Get("Content-Type"); got != "text/event-stream" {
		t.Fatalf("unexpected content-type header, got %s", got)
	}

	if body := recorder.Body.String(); body != expected {
		t.Fatalf("unexpected body, got %q", body)
	}
}
