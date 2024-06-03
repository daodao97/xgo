package xhttp

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

func WithValue(r *http.Request, key string, val any) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), key, val))
}

func WithRequestId(r *http.Request) *http.Request {
	return WithValue(r, "request_id", uuid.NewString())
}
