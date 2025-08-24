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

func (r *Response) ToHttpResponseWriter(w http.ResponseWriter, hooks ...ResponseHook) (int64, error) {
	var totalBytes int64

	w.WriteHeader(r.statusCode)
	for k, v := range r.RawResponse.Header {
		w.Header()[k] = v
	}
	if strings.Contains(r.RawResponse.Header.Get("Content-Type"), "text/event-stream") {
		reader := bufio.NewReader(r.RawResponse.Body)
		
		// 尝试读取第一行来判断是否为真正的 SSE 格式
		firstLine, err := reader.Peek(1024) // 预读取更多数据以判断格式
		if err != nil && err != io.EOF {
			return totalBytes, fmt.Errorf("error peeking response: %v", err)
		}
		
		// 检查是否包含换行符，如果没有换行符可能是不规范的单一响应
		if !bytes.Contains(firstLine, []byte("\n")) && err == io.EOF {
			// 处理不规范的响应，修正 Content-Type 为 application/json
			r.RawResponse.Header.Set("Content-Type", "application/json")
			w.Header().Set("Content-Type", "application/json")
			
			// 读取所有数据
			allData, readErr := io.ReadAll(reader)
			if readErr != nil {
				return totalBytes, fmt.Errorf("error reading non-standard response: %v", readErr)
			}
			
			// 应用 hooks
			flush := true
			processedData := allData
			for _, f := range hooks {
				flush, processedData = f(processedData)
			}
			
			if flush {
				n, writeErr := w.Write(processedData)
				if writeErr != nil {
					return totalBytes, fmt.Errorf("error writing response: %v", writeErr)
				}
				totalBytes += int64(n)
			}
			
			return totalBytes, nil
		}
		
		// 标准 SSE 流处理
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					return totalBytes, fmt.Errorf("error streaming response: %v", err)
				}
				// 如果遇到 EOF 但还有剩余数据，处理最后一行
				if len(line) > 0 {
					flush := true
					processedLine := line
					for _, f := range hooks {
						flush, processedLine = f(processedLine)
					}
					if flush {
						n, err := w.Write(processedLine)
						if err != nil {
							return totalBytes, fmt.Errorf("error writing final line: %v", err)
						}
						totalBytes += int64(n)
					}
				}
				return totalBytes, nil
			}

			// 保存原始的换行符
			originalLine := make([]byte, len(line))
			copy(originalLine, line)

			// 只去除右侧的换行符用于处理，保留其他空白字符
			trimmedLine := bytes.TrimRight(line, "\n")

			if len(trimmedLine) == 0 {
				// 如果是空行，直接写入原始内容
				n, err := w.Write(originalLine)
				if err != nil {
					return totalBytes, fmt.Errorf("error writing response: %v", err)
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

				// 恢复原始的换行符
				if bytes.HasSuffix(originalLine, []byte("\n")) {
					processedLine = append(processedLine, '\n')
				}

				// 写入响应，保持原有换行符
				n, err := w.Write(processedLine)
				if err != nil {
					return totalBytes, fmt.Errorf("error writing response: %v", err)
				}
				totalBytes += int64(n)
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
		n, err := w.Write(r.body)
		if err != nil {
			return totalBytes, fmt.Errorf("error writing response: %v", err)
		}
		totalBytes += int64(n)
	} else {
		if len(hooks) > 0 {
			body, _ := io.ReadAll(r.RawResponse.Body)
			for _, f := range hooks {
				_, body = f(body)
			}
			r.RawResponse.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		n, err := io.Copy(w, r.RawResponse.Body)
		if err != nil {
			return totalBytes, fmt.Errorf("error copying response: %v", err)
		}
		totalBytes += n
	}

	return totalBytes, nil
}
