package xrequest

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/daodao97/xgo/xjson"
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
	t.Log(resp.JSON())
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

func TestRequestWithReqHook(t *testing.T) {
	req := New().
		SetMethod(http.MethodPost).
		SetURL("http://127.0.0.1:8000").
		AddReqHook(func(req *http.Request) {
			req.Header.Add("X-Test", "test")
		})
	resp, err := req.Do()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp.StatusCode())
	t.Log(resp.JSON())
}
