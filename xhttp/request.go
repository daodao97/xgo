package xhttp

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

func Vars(r *http.Request, key string) string {
	return mux.Vars(r)[key]
}

func VarsDefault(r *http.Request, key string, defaultVal string) string {
	val, ok := mux.Vars(r)[key]
	if !ok {
		return defaultVal
	}
	return val
}

func Query(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

func QueryDefault(r *http.Request, key string, defaultVal string) string {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func ReadRequestBody(req *http.Request) ([]byte, error) {
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body.Close()
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return bodyBytes, nil
}

func DecodeBody[T any](r *http.Request) (*T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return nil, err
	}
	return &v, nil
}

// Decoder is a reusable schema decoder for form data, it's safe for concurrent use.
var decoder = schema.NewDecoder()

func DecodeFormData[T any](r *http.Request) (*T, error) {
	// Parse form data; this needs to be done before accessing form data
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	// Initialize the variable to store the decoded data
	var v T

	// Decode the form values into 'v'
	if err := decoder.Decode(&v, r.PostForm); err != nil {
		return nil, err
	}

	// Return the pointer to 'v'
	return &v, nil
}
