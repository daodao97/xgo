package xrequest

import (
	"net/http"
	"os"
	"testing"
	"time"
)

func TestRequest(t *testing.T) {

	b := `{
		"name": "daodao"
	}`

	request := New().
		SetMethod("POST").
		SetHeaders(map[string]string{
			"Content-Type": "application/json",
			"X-Test":       "test",
		}).
		SetBody(b).
		SetURL("https://httpbin.org/post").
		SetRetry(3, time.Second)
	resp, err := request.Do()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(resp.StatusCode())
	t.Log(resp.JSON())
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
	t.Log(resp.JSON())
}
