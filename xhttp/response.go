package xhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
)

func ResponseJson(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func ResponseErr(w http.ResponseWriter, err error, code int) {
	http.Error(w, err.Error(), code)
}

func ResponseSSE(w http.ResponseWriter, dataChan <-chan string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// 从通道中读取数据并发送给客户端
	for data := range dataChan {
		_, err := fmt.Fprintf(w, "data: %s\n\n", data)
		if err != nil {
			fmt.Println("error writing data:", err)
			return
		}
		flusher.Flush()
	}
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

func DecodeFormData[T any](r *http.Request) (*T, error) {
	var v T

	// Parse form data from the request
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("error parsing form: %w", err)
	}

	// Use reflection to fill in the fields of the struct v
	val := reflect.ValueOf(&v).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if field.CanSet() {
			// Get the form tag for the field
			tag := typ.Field(i).Tag.Get("form")
			if tag == "" {
				tag = typ.Field(i).Name // Fallback to field name
			}

			// Check if the value is provided in the form
			if values, exists := r.Form[tag]; exists && len(values) > 0 {
				// Assuming all fields are of string type for simplicity
				// Convert the first value from the form to the field's type
				field.Set(reflect.ValueOf(values[0]))
			}
		}
	}

	return &v, nil
}
