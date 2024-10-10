package xrequest

import (
	"testing"
	"time"
)

func TestRequest(t *testing.T) {
	request := New().
		SetMethod("GET").
		SetHeaders(map[string]string{"Content-Type": "application/json"}).
		SetURL("https://httpbin.org/get").
		SetDebug(true).
		SetRetry(3, time.Second)
	resp, err := request.Do()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(resp.StatusCode())
	t.Log(resp.String())
	t.Log(resp.JSON())
}
