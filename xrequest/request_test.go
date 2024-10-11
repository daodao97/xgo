package xrequest

import (
	"testing"
	"time"
)

func TestRequest(t *testing.T) {
	request := New().
		SetMethod("POST").
		SetHeaders(map[string]string{
			"Content-Type": "application/json",
			"X-Test":       "test",
		}).
		SetBody(map[string]any{
			"name": "daodao",
		}).
		SetURL("https://httpbin.org/post").
		SetRetry(3, time.Second)
	resp, err := request.Do()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(resp.StatusCode())
	t.Log(resp.String())
	t.Log(resp.JSON())
}
