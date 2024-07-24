package xresty

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

func NewRequest(req *http.Request) *resty.Request {
	r := resty.New()
	r.SetRetryCount(3)                     // 设置重试次数为 3
	r.SetRetryWaitTime(1 * time.Second)    // 设置每次重试间等待 1 秒
	r.SetRetryMaxWaitTime(2 * time.Second) // 设置最大重试间等待时间为 2 秒
	client := r.R()

	// 复制 cookies
	for _, c := range req.Cookies() {
		client.SetCookie(&http.Cookie{
			Name:  c.Name,
			Value: c.Value,
		})
	}

	// 复制 headers
	for k, v := range req.Header {
		client.SetHeader(k, v[0])
	}

	// 复制 query
	client.SetQueryParamsFromValues(req.URL.Query())

	// 复制 body
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err == nil {
			// 由于 req.Body 已被读取，需要重新赋值以备将来使用
			req.Body = io.NopCloser(bytes.NewBuffer(body))

			client.SetBody(body)
		}
	}

	return client
}

// ResponseToHTTPResponse 将 resty.Response 重写到 http.ResponseWriter
func ResponseToHTTPResponse(w http.ResponseWriter, resp *resty.Response) error {
	// 将 Resty 响应状态码写入 http.ResponseWriter
	w.WriteHeader(resp.StatusCode())

	// 将 Resty 响应头部写入 http.ResponseWriter
	for key, values := range resp.Header() {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// 将 Resty 响应体写入 http.ResponseWriter
	_, err := io.Copy(w, resp.RawBody())
	return err
}
