package xrequest

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	"github.com/daodao97/xgo/xjson"
)

func NewResponse(rawResponse *http.Response, parseResponse bool) *Response {
	var body []byte
	if parseResponse {
		body, _ = io.ReadAll(rawResponse.Body)
		rawResponse.Body.Close()
		rawResponse.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	return &Response{RawResponse: rawResponse, statusCode: rawResponse.StatusCode, body: body}
}

type Response struct {
	RawResponse *http.Response
	statusCode  int
	body        []byte
}

func (r *Response) StatusCode() int {
	return r.statusCode
}

func (r *Response) String() string {
	return string(r.body)
}

func (r *Response) JSON() *xjson.Json {
	return xjson.New(r.body)
}

func (r *Response) Scan(dest any) error {
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
	return r.body
}

func (r *Response) XML(v any) error {
	return xml.Unmarshal(r.body, v)
}

func (r *Response) Headers() http.Header {
	return r.RawResponse.Header
}
