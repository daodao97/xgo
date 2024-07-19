package xhttp

import (
	"encoding/json"
	"fmt"
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
