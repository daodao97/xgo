package xrequest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/daodao97/xgo/utils"
)

type Request struct {
	mu            sync.Mutex
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
	queryParams   map[string]string

	// auth
	basicAuth bool
	username  string
	password  string

	// retry
	retryAttempts uint
	retryDelay    time.Duration

	// client
	client *http.Client
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

// 是否解析响应体
// 默认解析, 如果设置为 false, 则需要自行关闭 body.close
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

func (r *Request) SetQueryParams(queryParams map[string]string) *Request {
	r.queryParams = queryParams
	return r
}

func (r *Request) SetRetry(attempts uint, delay time.Duration) *Request {
	r.retryAttempts = attempts
	r.retryDelay = delay
	return r
}

func (r *Request) SetClient(client *http.Client) *Request {
	r.client = client
	return r
}

func (r *Request) Get(targetUrl string) (resp *Response, err error) {
	return r.SetMethod(http.MethodGet).SetURL(targetUrl).Do()
}

func (r *Request) Post(targetUrl string) (resp *Response, err error) {
	return r.SetMethod(http.MethodPost).SetURL(targetUrl).Do()
}

func (r *Request) Put(targetUrl string) (resp *Response, err error) {
	return r.SetMethod(http.MethodPut).SetURL(targetUrl).Do()
}

func (r *Request) Delete(targetUrl string) (resp *Response, err error) {
	return r.SetMethod(http.MethodDelete).SetURL(targetUrl).Do()
}

func (r *Request) Patch(targetUrl string) (resp *Response, err error) {
	return r.SetMethod(http.MethodPatch).SetURL(targetUrl).Do()
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

	body, err := r.prepareBody()
	if err != nil {
		return nil, err
	}

	if r.queryParams != nil {
		parsedURL, err := url.Parse(targetUrl)
		if err != nil {
			return nil, NewRequestError("解析 URL 失败", err)
		}

		queryParams := parsedURL.Query()
		for k, v := range r.queryParams {
			queryParams.Add(k, v)
		}

		parsedURL.RawQuery = queryParams.Encode()
		targetUrl = parsedURL.String()
	}

	req, err := http.NewRequest(method, targetUrl, body)
	if err != nil {
		return nil, NewRequestError("创建请求失败", err)
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

	client := r.client
	if client == nil {
		client = &http.Client{}
	}
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
		return nil, NewRequestError("请求失败", err)
	}

	return NewResponse(resp, r.parseResponse), nil
}

func (r *Request) prepareBody() (io.Reader, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.body != nil {
		jsonBody, err := json.Marshal(r.body)
		if err != nil {
			return nil, NewRequestError("序列化请求数据失败", err)
		}
		return bytes.NewBuffer(jsonBody), nil
	}

	if r.formData != nil || r.formUrlEncode != nil {
		formBody := url.Values{}
		data := r.formData
		if r.formUrlEncode != nil {
			data = r.formUrlEncode
		}
		for k, v := range data {
			formBody.Add(k, fmt.Sprintf("%v", v))
		}
		r.headers["Content-Type"] = "application/x-www-form-urlencoded"
		return strings.NewReader(formBody.Encode()), nil
	}

	return nil, nil
}
