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

	return &Response{RawResponse: rawResponse, StatusCode: rawResponse.StatusCode, Body: body}
}

type Response struct {
	RawResponse *http.Response
	StatusCode  int
	Body        []byte
}

func (r *Response) String() string {
	return string(r.Body)
}

func (r *Response) JSON() *xjson.Json {
	return xjson.New(r.Body)
}

func (r *Response) Scan(dest any) error {
	return json.Unmarshal(r.Body, dest)
}

func (r *Response) IsError() bool {
	return r.StatusCode > 299
}
