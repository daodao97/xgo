package xhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/schema"
	"io"
	"net/http"
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
