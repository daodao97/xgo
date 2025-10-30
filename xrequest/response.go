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
	isStream    bool
}

func (r *Response) BodyIsEmpty() bool {
	// 如果已经解析，检查解析后的内容
	if r.parsed {
		if len(r.body) == 0 {
			return true
		}
		return len(strings.TrimSpace(string(r.body))) == 0
	}

	// 未解析时，根据 Content-Length 判断
	contentLength := r.RawResponse.ContentLength
	if contentLength == 0 {
		return true
	}

	// Content-Length 为 -1 表示未知长度，需要 peek 判断
	if contentLength == -1 {
		if r.RawResponse.Body == nil {
			return true
		}

		// 使用 peek 检查是否有数据，不消耗原始 body
		peekReader := bufio.NewReader(r.RawResponse.Body)
		_, err := peekReader.Peek(1)
		if err == io.EOF {
			return true
		}
		// 将 peekReader 重新包装回 RawResponse.Body
		r.RawResponse.Body = io.NopCloser(peekReader)
		return false
	}

	// Content-Length > 0，但可能内容全是空白字符
	// 对于 SSE 这种流式响应，也需要能判断是否真的有内容
	if contentLength > 0 {
		return false
	}

	return false
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

func (r *Response) Stream() (chan string, error) {
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

type ResponseHook func(data []byte) (flush bool, newData []byte)

func copyResponseHeaders(dst, src http.Header) {
	if dst == nil || src == nil {
		return
	}
	for key, values := range src {
		copied := make([]string, len(values))
		copy(copied, values)
		dst[key] = copied
	}
}

func (r *Response) ToHttpResponseWriterWihtStream(w http.ResponseWriter, isStream bool, hooks ...ResponseHook) (int64, error) {
	r.isStream = isStream

	return r.ToHttpResponseWriter(w, hooks...)
}

func (r *Response) ToHttpResponseWriter(w http.ResponseWriter, hooks ...ResponseHook) (int64, error) {
	if r.RawResponse == nil {
		return 0, fmt.Errorf("raw response is nil")
	}

	body := r.RawResponse.Body
	if body != nil {
		defer body.Close()
	}

	statusCode := r.StatusCode()
	contentType := r.RawResponse.Header.Get("Content-Type")

	writeHeaders := func() {
		copyResponseHeaders(w.Header(), r.RawResponse.Header)
		w.WriteHeader(statusCode)
	}

	if statusCode >= http.StatusBadRequest {
		writeHeaders()
		return r.notStreamResponse(w, hooks...)
	}

	if (strings.Contains(contentType, "text/event-stream") || r.isStream) && body != nil {
		reader := bufio.NewReader(body)

		firstLine, err := reader.Peek(1024)
		if err != nil && err != io.EOF {
			return 0, fmt.Errorf("error peeking response: %w", err)
		}

		if !bytes.Contains(firstLine, []byte("\n")) && err == io.EOF {
			r.RawResponse.Header.Set("Content-Type", "application/json")
			writeHeaders()

			allData, readErr := io.ReadAll(reader)
			if readErr != nil {
				return 0, fmt.Errorf("error reading non-standard response: %w", readErr)
			}

			totalBytes := int64(0)
			flush := true
			processedData := allData
			for _, f := range hooks {
				flush, processedData = f(processedData)
			}

			if flush && len(processedData) > 0 {
				n, writeErr := w.Write(processedData)
				if writeErr != nil {
					return totalBytes, fmt.Errorf("error writing response: %w", writeErr)
				}
				totalBytes += int64(n)
			}

			return totalBytes, nil
		}

		writeHeaders()

		totalBytes := int64(0)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					return totalBytes, fmt.Errorf("error streaming response: %w", err)
				}
				if len(line) > 0 {
					flush := true
					processedLine := line
					for _, f := range hooks {
						flush, processedLine = f(processedLine)
					}
					if flush {
						n, writeErr := w.Write(processedLine)
						if writeErr != nil {
							return totalBytes, fmt.Errorf("error writing final line: %w", writeErr)
						}
						totalBytes += int64(n)
					}
				}
				return totalBytes, nil
			}

			originalLine := make([]byte, len(line))
			copy(originalLine, line)

			trimmedLine := bytes.TrimRight(line, "\n")

			if len(trimmedLine) == 0 {
				n, writeErr := w.Write(originalLine)
				if writeErr != nil {
					return totalBytes, fmt.Errorf("error writing response: %w", writeErr)
				}
				totalBytes += int64(n)
			} else {
				flush := true
				processedLine := trimmedLine
				for _, f := range hooks {
					flush, processedLine = f(processedLine)
				}
				if !flush {
					continue
				}

				if bytes.HasSuffix(originalLine, []byte("\n")) {
					processedLine = append(processedLine, '\n')
				}

				n, writeErr := w.Write(processedLine)
				if writeErr != nil {
					return totalBytes, fmt.Errorf("error writing response: %w", writeErr)
				}
				totalBytes += int64(n)
			}

			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}

	writeHeaders()
	return r.notStreamResponse(w, hooks...)
}

func (r *Response) notStreamResponse(w http.ResponseWriter, hooks ...ResponseHook) (int64, error) {
	totalBytes := int64(0)
	if r.parsed {
		for _, f := range hooks {
			_, r.body = f(r.body)
		}
		n, err := w.Write(r.body)
		if err != nil {
			return totalBytes, fmt.Errorf("error writing response: %w", err)
		}
		totalBytes += int64(n)
	} else {
		if r.RawResponse.Body == nil {
			return totalBytes, nil
		}
		if len(hooks) > 0 {
			body, _ := io.ReadAll(r.RawResponse.Body)
			for _, f := range hooks {
				_, body = f(body)
			}
			r.RawResponse.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		n, err := io.Copy(w, r.RawResponse.Body)
		if err != nil {
			return totalBytes, fmt.Errorf("error copying response: %w", err)
		}
		totalBytes += n
	}

	return totalBytes, nil
}
