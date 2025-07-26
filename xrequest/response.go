package xrequest

import (
	"bufio"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/daodao97/xgo/xcode"
	"github.com/daodao97/xgo/xjson"
)

func NewResponse(rawResponse *http.Response) *Response {
	return &Response{RawResponse: rawResponse, statusCode: rawResponse.StatusCode}
}

type Response struct {
	RawResponse *http.Response
	statusCode  int
	body        []byte
	parsed      bool
}

func (r *Response) parseResponse() bool {
	if r.parsed {
		return true
	}
	if strings.Contains(r.RawResponse.Header.Get("Content-Type"), "text/event-stream") {
		return false
	}
	
	// 获取原始响应体
	var reader io.Reader = r.RawResponse.Body
	
	// 检查Content-Encoding并处理压缩
	encoding := r.RawResponse.Header.Get("Content-Encoding")
	switch encoding {
	case "gzip":
		if gzipReader, err := gzip.NewReader(r.RawResponse.Body); err == nil {
			reader = gzipReader
			defer gzipReader.Close()
			// 删除Content-Encoding头，因为内容已解压缩
			r.RawResponse.Header.Del("Content-Encoding")
		}
	case "deflate":
		reader = flate.NewReader(r.RawResponse.Body)
		defer reader.(io.ReadCloser).Close()
		// 删除Content-Encoding头，因为内容已解压缩
		r.RawResponse.Header.Del("Content-Encoding")
	case "br":
		reader = brotli.NewReader(r.RawResponse.Body)
		// 删除Content-Encoding头，因为内容已解压缩
		r.RawResponse.Header.Del("Content-Encoding")
	}
	
	var body []byte
	body, _ = io.ReadAll(reader)
	r.RawResponse.Body.Close()
	r.RawResponse.Body = io.NopCloser(bytes.NewBuffer(body))
	r.parsed = true
	r.body = body
	return true
}

func (r *Response) StatusCode() int {
	return r.statusCode
}

func (r *Response) String() string {
	r.parseResponse()
	return string(r.body)
}

// Deprecated: use Json instead
func (r *Response) JSON() *xjson.Json {
	r.parseResponse()
	return xjson.New(r.body)
}

func (r *Response) Json() *xjson.Json {
	r.parseResponse()
	return xjson.New(r.body)
}

func (r *Response) Scan(dest any) error {
	r.parseResponse()
	return json.Unmarshal(r.body, dest)
}

func (r *Response) IsError() bool {
	return r.statusCode >= http.StatusBadRequest
}

func (r *Response) Error() error {
	if r.IsError() {
		return errors.New(string(r.body))
	}
	return nil
}

func (r *Response) Bytes() []byte {
	r.parseResponse()
	return r.body
}

func (r *Response) XML(v any) error {
	r.parseResponse()
	return xml.Unmarshal(r.body, v)
}

func (r *Response) Headers() http.Header {
	return r.RawResponse.Header
}

func (r *Response) SSE() (chan string, error) {
	if !strings.Contains(r.RawResponse.Header.Get("Content-Type"), "text/event-stream") {
		return nil, &xcode.Code{
			HttpCode: http.StatusBadRequest,
			Message:  r.String(),
			Type:     "error",
		}
	}
	messages := make(chan string)

	go func() {
		defer close(messages)
		reader := bufio.NewReader(r.RawResponse.Body)

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					fmt.Printf("Error reading SSE stream: %v\n", err)
				}
				return
			}

			messages <- line

			// line = strings.TrimSpace(line)
			// if strings.HasPrefix(line, "data: ") {
			// 	data := strings.TrimPrefix(line, "data: ")
			// 	messages <- data
			// } else {
			// 	messages <- line
			// }
		}
	}()

	return messages, nil
}

func (r *Response) Stream() (chan string, error) {
	return r.SSE()
}

type ResponseHook func(data []byte) (flush bool, newData []byte)

func (r *Response) ToHttpResponseWriter(w http.ResponseWriter, hooks ...ResponseHook) {
	w.WriteHeader(r.statusCode)
	for k, v := range r.RawResponse.Header {
		w.Header()[k] = v
	}
	if strings.Contains(r.RawResponse.Header.Get("Content-Type"), "text/event-stream") {
		reader := bufio.NewReader(r.RawResponse.Body)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					http.Error(w, fmt.Sprintf("Error streaming response: %v", err), http.StatusInternalServerError)
				}
				return
			}

			// 保存原始的换行符
			originalLine := make([]byte, len(line))
			copy(originalLine, line)

			// 只去除右侧的换行符用于处理，保留其他空白字符
			trimmedLine := bytes.TrimRight(line, "\n")

			if len(trimmedLine) == 0 {
				// 如果是空行，直接写入原始内容
				if _, err := w.Write(originalLine); err != nil {
					http.Error(w, fmt.Sprintf("Error writing response: %v", err), http.StatusInternalServerError)
					return
				}
			} else {
				flush := true
				processedLine := trimmedLine
				for _, f := range hooks {
					flush, processedLine = f(processedLine)
				}
				if !flush {
					continue
				}

				// 恢复原始的换行符
				if bytes.HasSuffix(originalLine, []byte("\n")) {
					processedLine = append(processedLine, '\n')
				}

				// 写入响应，保持原有换行符
				if _, err := w.Write(processedLine); err != nil {
					http.Error(w, fmt.Sprintf("Error writing response: %v", err), http.StatusInternalServerError)
					return
				}
			}

			// 刷新响应
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}

	if r.parsed {
		for _, f := range hooks {
			_, r.body = f(r.body)
		}
		if _, err := w.Write(r.body); err != nil {
			http.Error(w, fmt.Sprintf("Error writing response: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		if len(hooks) > 0 {
			body, _ := io.ReadAll(r.RawResponse.Body)
			for _, f := range hooks {
				_, body = f(body)
			}
			r.RawResponse.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		if _, err := io.Copy(w, r.RawResponse.Body); err != nil {
			http.Error(w, fmt.Sprintf("Error copying response: %v", err), http.StatusInternalServerError)
			return
		}
	}
}
