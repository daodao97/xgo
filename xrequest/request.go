package xrequest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/daodao97/xgo/utils"
)

type Request struct {
	debug         bool
	method        string
	targetUrl     string
	body          any
	parseResponse bool
	headers       map[string]string
	proxy         string
	timeout       time.Duration
	formData      map[string]any
	formUrlEncode map[string]any

	// auth
	basicAuth bool
	username  string
	password  string

	// retry
	retryAttempts uint
	retryDelay    time.Duration
}

func New() *Request {
	return &Request{parseResponse: true, debug: utils.IsGoRun()}
}

func (r *Request) SetMethod(method string) *Request {
	r.method = method
	return r
}

func (r *Request) SetURL(url string) *Request {
	r.targetUrl = url
	return r
}

func (r *Request) SetBody(body any) *Request {
	r.body = body
	return r
}

func (r *Request) SetDebug(debug bool) *Request {
	r.debug = debug
	return r
}

func (r *Request) SetParseResponse(parseResponse bool) *Request {
	r.parseResponse = parseResponse
	return r
}

func (r *Request) SetHeaders(headers map[string]string) *Request {
	r.headers = headers
	return r
}

func (r *Request) SetHeader(key, value string) *Request {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}
	r.headers[key] = value
	return r
}

func (r *Request) SetProxy(proxy string) *Request {
	r.proxy = proxy
	return r
}

func (r *Request) SetTimeout(timeout time.Duration) *Request {
	r.timeout = timeout
	return r
}

func (r *Request) SetBasicAuth(username, password string) *Request {
	r.basicAuth = true
	r.username = username
	r.password = password
	return r
}

func (r *Request) SetFormData(formData map[string]any) *Request {
	r.formData = formData
	return r
}

func (r *Request) SetFormUrlEncode(formUrlEncode map[string]any) *Request {
	r.formUrlEncode = formUrlEncode
	return r
}

func (r *Request) SetRetry(attempts uint, delay time.Duration) *Request {
	r.retryAttempts = attempts
	r.retryDelay = delay
	return r
}

func (r *Request) Do() (resp *Response, err error) {
	if r.retryAttempts == 0 {
		return r.do()
	}

	retry.Do(func() error {
		resp, err = r.do()
		return err
	}, retry.Attempts(r.retryAttempts), retry.Delay(r.retryDelay))

	return
}

func (r *Request) do() (*Response, error) {
	method := r.method
	if method == "" {
		method = http.MethodGet
	}
	targetUrl := r.targetUrl

	var body io.Reader
	if r.body != nil {
		jsonBody, err := json.Marshal(r.body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求数据失败: %w", err)
		}
		body = bytes.NewBuffer(jsonBody)
	} else {
		body = nil
	}

	if r.formData != nil {
		formBody := url.Values{}
		for k, v := range r.formData {
			formBody.Add(k, fmt.Sprintf("%v", v))
		}
		body = strings.NewReader(formBody.Encode())
		r.headers["Content-Type"] = "application/x-www-form-urlencoded"
	}

	if r.formUrlEncode != nil {
		formBody := url.Values{}
		for k, v := range r.formUrlEncode {
			formBody.Add(k, fmt.Sprintf("%v", v))
		}
		body = strings.NewReader(formBody.Encode())
		r.headers["Content-Type"] = "application/x-www-form-urlencoded"
	}

	req, err := http.NewRequest(method, targetUrl, body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	if r.basicAuth {
		req.SetBasicAuth(r.username, r.password)
	}

	if r.debug {
		_curl, _ := GetCurlCommand(req)
		fmt.Println(_curl)
	}

	client := &http.Client{}
	if r.proxy != "" {
		client.Transport = &http.Transport{Proxy: func(_ *http.Request) (*url.URL, error) {
			return url.Parse(r.proxy)
		}}
	}
	if r.timeout > 0 {
		client.Timeout = r.timeout
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("xrequest failed: %w", err)
	}

	return NewResponse(resp, r.parseResponse), nil
}
