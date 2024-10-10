package xrequest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/daodao97/xgo/xjson"
)

func NewResponse(rawResponse *http.Response, parseResponse bool) *Response {
	var body []byte
	if parseResponse {
		body, _ = io.ReadAll(rawResponse.Body)
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
