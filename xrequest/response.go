package xrequest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/daodao97/xgo/xjson"
)

func NewResponse(rawResponse *http.Response, parseResponse bool) *Response {
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
	if r.RawResponse.Header.Get("Content-Type") == "text/event-stream" {
		return false
	}
	var body []byte
	body, _ = io.ReadAll(r.RawResponse.Body)
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

// deprecated
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
		return fmt.Errorf("request failed, status code: %d, body: %s", r.statusCode, string(r.body))
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
	if r.RawResponse.Header.Get("Content-Type") != "text/event-stream" {
		return nil, fmt.Errorf("response is not an SSE")
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

			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				messages <- data
			}
		}
	}()

	return messages, nil
}
