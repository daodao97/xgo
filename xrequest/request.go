package xrequest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/daodao97/xgo/xlog"
)

var RequestDebug = false

func SetRequestDebug(debug bool) {
	RequestDebug = debug
}

type Request struct {
	mu            sync.Mutex
	debug         bool
	method        string
	targetUrl     string
	body          any
	headers       map[string]string
	cookies       map[string]string
	proxy         string
	timeout       time.Duration
	formData      map[string]string
	formUrlEncode map[string]string
	queryParams   map[string]string
	files         map[string][]File
	ctx           context.Context
	req           *http.Request

	// auth
	basicAuth bool
	username  string
	password  string

	// retry
	retryAttempts  uint
	retryDelay     time.Duration
	retryCondition func(*Response, error) bool

	// client
	client *http.Client

	reqHooks []func(req *http.Request) error
}

type File struct {
	FieldName string
	FileName  string
	Content   io.Reader
}

func New() *Request {
	return &Request{}
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

func (r *Request) SetHeaders(headers map[string]string) *Request {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}
	for key, value := range headers {
		r.headers[key] = value
	}
	return r
}

func (r *Request) SetHeader(key, value string) *Request {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}
	r.headers[key] = value
	return r
}

func (r *Request) SetCookies(cookies map[string]string) *Request {
	if r.cookies == nil {
		r.cookies = make(map[string]string)
	}
	for key, value := range cookies {
		r.cookies[key] = value
	}
	return r
}

func (r *Request) SetCookie(key, value string) *Request {
	if r.cookies == nil {
		r.cookies = make(map[string]string)
	}
	r.cookies[key] = value
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

func (r *Request) SetFormData(formData map[string]string) *Request {
	r.formData = formData
	return r
}

func (r *Request) SetFormUrlEncode(formUrlEncode map[string]string) *Request {
	r.formUrlEncode = formUrlEncode
	return r
}

func (r *Request) SetQueryParams(queryParams map[string]string) *Request {
	if r.queryParams == nil {
		r.queryParams = make(map[string]string)
	}
	for key, value := range queryParams {
		r.queryParams[key] = value
	}
	return r
}

func (r *Request) SetQueryParam(key, value string) *Request {
	if r.queryParams == nil {
		r.queryParams = make(map[string]string)
	}
	r.queryParams[key] = value
	return r
}

func (r *Request) SetRetry(attempts uint, delay time.Duration) *Request {
	r.retryAttempts = attempts
	r.retryDelay = delay
	return r
}

func (r *Request) SetRetryWithCondition(attempts uint, delay time.Duration, condition func(*Response, error) bool) *Request {
	r.retryAttempts = attempts
	r.retryDelay = delay
	r.retryCondition = condition
	return r
}

func (r *Request) SetClient(client *http.Client) *Request {
	r.client = client
	return r
}

func (r *Request) AddFile(fieldName, fileName string, content io.Reader) *Request {
	if r.files == nil {
		r.files = make(map[string][]File)
	}
	r.files[fieldName] = append(r.files[fieldName], File{
		FieldName: fieldName,
		FileName:  fileName,
		Content:   content,
	})
	return r
}

func (r *Request) AddReqHook(hook func(req *http.Request) error) *Request {
	r.reqHooks = append(r.reqHooks, hook)
	return r
}

func (r *Request) WithContext(ctx context.Context) *Request {
	r.ctx = ctx
	return r
}

func (r *Request) WithRequest(req *http.Request) *Request {
	r.req = req
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

	var lastResp *Response
	var lastErr error
	for attempt := uint(1); attempt <= r.retryAttempts; attempt++ {
		resp, err = r.do()

		if !r.shouldRetry(resp, err) {
			return resp, err
		}

		lastResp = resp
		lastErr = err
		if attempt < r.retryAttempts {
			time.Sleep(r.retryDelay)
		}
	}

	return lastResp, lastErr
}

func (r *Request) shouldRetry(resp *Response, err error) bool {
	if r.retryCondition != nil {
		return r.retryCondition(resp, err)
	}

	// 默认重试逻辑：网络错误或服务器错误(>= 500)
	if err != nil {
		return true
	}

	return resp != nil && resp.StatusCode() >= http.StatusInternalServerError
}

func (r *Request) do() (*Response, error) {
	ctx := r.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	r.debug = RequestDebug

	req, err := r.makeRequest(ctx)
	if err != nil {
		return nil, NewRequestError("创建请求失败", err)
	}

	var debugInfo []string
	var _curl *CurlCommand
	var _curlString string

	if _curl, err = GetCurlCommand(req); err == nil {
		debugInfo = append(debugInfo, _curl.String())
		_curlString = _curl.String()
	} else {
		xlog.WarnCtx(ctx, "getCurlCommand error", xlog.Any("error", err))
	}

	client := r.client
	if client == nil {
		client = GetDefaultProxyClient()
	}
	if r.proxy != "" {
		if u, err := url.Parse(r.proxy); err == nil {
			client.Transport = &http.Transport{Proxy: proxyWithNoProxy(u)}
		}
	}
	if r.timeout > 0 {
		client.Timeout = r.timeout
	}

	start := time.Now()
	resp, err := client.Do(req)

	duration := time.Since(start)
	logFunc := xlog.DebugCtx
	args := []any{
		xlog.String("url", r.targetUrl),
		xlog.String("method", r.method),
		xlog.Duration("duration", duration),
		xlog.Any("curl", _curlString),
	}

	_ = _curlString

	if err != nil {
		logFunc = xlog.WarnCtx
		args = append(args, xlog.Any("error", err))
		logFunc(ctx, "xrequest network error", args...)
		if r.debug && len(debugInfo) > 0 {
			debugInfo = append([]string{"-------request curl command start-------"}, debugInfo...)

			debugInfo = append(debugInfo, "-------request curl command end-------")
			fmt.Println(strings.Join(debugInfo, "\n"))
		}

		return nil, NewRequestError("请求失败", err)
	}

	if resp != nil {
		args = append(args, xlog.Any("status", resp.StatusCode))
	}

	_resp := NewResponse(resp)
	if r.debug && len(debugInfo) > 0 {
		debugInfo = append([]string{"-------request curl command start-------"}, debugInfo...)
		debugInfo = append(debugInfo, fmt.Sprintf("response status: %d", resp.StatusCode))

		// 只在响应是文本类型时打印响应体
		contentType := resp.Header.Get("Content-Type")
		if strings.Contains(contentType, "text") ||
			strings.Contains(contentType, "json") ||
			strings.Contains(contentType, "xml") {
			debugInfo = append(debugInfo, fmt.Sprintf("response body: %s", _resp.String()))
		}

		debugInfo = append(debugInfo, "-------request curl command end-------")
		fmt.Println(strings.Join(debugInfo, "\n"))
	}

	if _resp.Error() != nil {
		logFunc = xlog.WarnCtx
		args = append(args, xlog.Any("error", _resp.Error()), xlog.String("response", _resp.String()))
	}

	logFunc(ctx, "xrequest", args...)

	return _resp, nil
}

func (r *Request) makeRequest(ctx context.Context) (*http.Request, error) {
	if r.req != nil {
		newReq := &http.Request{
			Method: r.req.Method,
			Header: r.req.Header.Clone(),
			Body:   r.req.Body,
		}

		_url, err := url.Parse(r.targetUrl)
		if err != nil {
			return nil, NewRequestError("解析 URL 失败", err)
		}

		if r.req.URL != nil {
			newReq.URL = r.req.URL.ResolveReference(_url)
		} else {
			newReq.URL = _url
		}

		return newReq, nil
	}

	method := r.method
	if method == "" {
		method = http.MethodGet
	}
	targetUrl := r.targetUrl

	body, err := r.prepareBody()
	if err != nil {
		return nil, NewRequestError("准备请求体失败", err)
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

	req, err := http.NewRequestWithContext(ctx, method, targetUrl, body)
	if err != nil {
		return nil, NewRequestError("创建请求失败", err)
	}

	for k, v := range r.headers {
		// req.Header[formatHeaderKey(k)] = []string{v}
		req.Header[k] = []string{v}
	}

	for k, v := range r.cookies {
		req.AddCookie(&http.Cookie{Name: k, Value: v})
	}

	if r.basicAuth {
		req.SetBasicAuth(r.username, r.password)
	}

	for _, hook := range r.reqHooks {
		if err := hook(req); err != nil {
			return nil, NewRequestError("请求钩子执行失败", err)
		}
	}

	return req, nil
}

func (r *Request) prepareBody() (io.Reader, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.headers == nil {
		r.headers = make(map[string]string)
	}

	if r.files != nil {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// 设置 Content-Type header 和 boundary
		if _, exists := r.headers["Content-Type"]; !exists {
			r.headers["Content-Type"] = writer.FormDataContentType()
		}

		// 处理文件上传
		for _, files := range r.files {
			for _, file := range files {
				part, err := writer.CreateFormFile(file.FieldName, file.FileName)
				if err != nil {
					return nil, NewRequestError("创建文件表单失败", err)
				}
				_, err = io.Copy(part, file.Content)
				if err != nil {
					return nil, NewRequestError("写入文件内容失败", err)
				}
			}
		}

		// 处理其他表单数据
		if r.formData != nil {
			for key, value := range r.formData {
				err := writer.WriteField(key, fmt.Sprintf("%v", value))
				if err != nil {
					return nil, NewRequestError("写入表单字段失败", err)
				}
			}
		}

		err := writer.Close()
		if err != nil {
			return nil, NewRequestError("关闭multipart writer失败", err)
		}

		return body, nil
	}

	if r.body != nil {
		if _, exists := r.headers["Content-Type"]; !exists {
			r.headers["Content-Type"] = "application/json"
		}
		switch v := r.body.(type) {
		case string:
			return strings.NewReader(v), nil
		case []byte:
			return bytes.NewReader(v), nil
		case io.Reader:
			return v, nil
		default:
			if stringer, ok := v.(fmt.Stringer); ok {
				return strings.NewReader(stringer.String()), nil
			}
			jsonBody, err := json.Marshal(r.body)
			if err != nil {
				return nil, NewRequestError("序列化请求数据失败", err)
			}
			return bytes.NewBuffer(jsonBody), nil
		}
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
		if _, exists := r.headers["Content-Type"]; !exists {
			r.headers["Content-Type"] = "application/x-www-form-urlencoded"
		}
		return strings.NewReader(formBody.Encode()), nil
	}

	return nil, nil
}

func formatHeaderKey(key string) string {
	parts := strings.Split(key, "-")
	for i, part := range parts {
		parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
	}
	return strings.Join(parts, "-")
}
