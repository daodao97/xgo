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
)

type Request struct {
	Debug         bool
	Method        string
	URL           string
	Body          any
	ParseResponse bool
	Headers       map[string]string
	Proxy         string
	Timeout       time.Duration
	FormData      map[string]any
	FormUrlEncode map[string]any

	// auth
	BasicAuth bool
	Username  string
	Password  string

	// retry
	RetryAttempts uint
	RetryDelay    time.Duration
}

func New() *Request {
	return &Request{ParseResponse: true}
}

func (r *Request) SetMethod(method string) *Request {
	r.Method = method
	return r
}

func (r *Request) SetURL(url string) *Request {
	r.URL = url
	return r
}

func (r *Request) SetBody(body any) *Request {
	r.Body = body
	return r
}

func (r *Request) SetDebug(debug bool) *Request {
	r.Debug = debug
	return r
}

func (r *Request) SetParseResponse(parseResponse bool) *Request {
	r.ParseResponse = parseResponse
	return r
}

func (r *Request) SetHeaders(headers map[string]string) *Request {
	r.Headers = headers
	return r
}

func (r *Request) SetHeader(key, value string) *Request {
	if r.Headers == nil {
		r.Headers = make(map[string]string)
	}
	r.Headers[key] = value
	return r
}

func (r *Request) SetProxy(proxy string) *Request {
	r.Proxy = proxy
	return r
}

func (r *Request) SetTimeout(timeout time.Duration) *Request {
	r.Timeout = timeout
	return r
}

func (r *Request) SetBasicAuth(username, password string) *Request {
	r.BasicAuth = true
	r.Username = username
	r.Password = password
	return r
}

func (r *Request) SetFormData(formData map[string]any) *Request {
	r.FormData = formData
	return r
}

func (r *Request) SetFormUrlEncode(formUrlEncode map[string]any) *Request {
	r.FormUrlEncode = formUrlEncode
	return r
}

func (r *Request) SetRetry(attempts uint, delay time.Duration) *Request {
	r.RetryAttempts = attempts
	r.RetryDelay = delay
	return r
}

func (r *Request) Do() (resp *Response, err error) {
	if r.RetryAttempts == 0 {
		return r.do()
	}

	retry.Do(func() error {
		resp, err = r.do()
		return err
	}, retry.Attempts(r.RetryAttempts), retry.Delay(r.RetryDelay))

	return
}

func (r *Request) do() (*Response, error) {
	method := r.Method
	if method == "" {
		method = http.MethodGet
	}
	targetUrl := r.URL

	var body io.Reader
	if r.Body != nil {
		jsonBody, err := json.Marshal(r.Body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求数据失败: %w", err)
		}
		body = bytes.NewBuffer(jsonBody)
	} else {
		body = nil
	}

	if r.FormData != nil {
		formBody := url.Values{}
		for k, v := range r.FormData {
			formBody.Add(k, fmt.Sprintf("%v", v))
		}
		body = strings.NewReader(formBody.Encode())
		r.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	}

	if r.FormUrlEncode != nil {
		formBody := url.Values{}
		for k, v := range r.FormUrlEncode {
			formBody.Add(k, fmt.Sprintf("%v", v))
		}
		body = strings.NewReader(formBody.Encode())
		r.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	}

	req, err := http.NewRequest(method, targetUrl, body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	for k, v := range r.Headers {
		req.Header.Set(k, v)
	}

	if r.BasicAuth {
		req.SetBasicAuth(r.Username, r.Password)
	}

	if r.Debug {
		_curl, _ := GetCurlCommand(req)
		fmt.Println(_curl)
	}

	client := &http.Client{}
	if r.Proxy != "" {
		client.Transport = &http.Transport{Proxy: func(_ *http.Request) (*url.URL, error) {
			return url.Parse(r.Proxy)
		}}
	}
	if r.Timeout > 0 {
		client.Timeout = r.Timeout
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("xrequest failed: %w", err)
	}

	return NewResponse(resp, r.ParseResponse), nil
}
