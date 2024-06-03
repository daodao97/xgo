package xhttp

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"github.com/daodao97/xgo/xlog"
	"log/slog"
	"net/http"
	"runtime"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func PanicRecovery(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				if _err, ok := err.(error); ok {
					if errors.Is(_err, http.ErrAbortHandler) {
						return
					}
				}
				buf := make([]byte, 2048)
				n := runtime.Stack(buf, false)
				buf = buf[:n]

				xlog.Error(fmt.Sprintf("recovering from %s err %v\n %s", r.URL.Path, err, buf))
				w.Write([]byte(`{"error":"something went wrong"}`))
			}
		}()

		h.ServeHTTP(w, r)
	})
}

var (
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "myapp_http_duration_seconds",
		Help: "Duration of HTTP requests.",
	}, []string{"path"})
)

// PrometheusMiddleware implements mux.MiddlewareFunc.
func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))
		next.ServeHTTP(w, r)
		timer.ObserveDuration()
	})
}

func BasicAuth(handler http.HandlerFunc, username, password, realm string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
			return
		}

		handler(w, r)
	}
}

func RequestLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		r = WithRequestId(r)

		// 创建一个响应记录器
		lr := &responseLogger{ResponseWriter: w, statusCode: http.StatusOK}

		// 调用下一个中间件或最终的处理器
		next.ServeHTTP(lr, r)

		duration := time.Since(start)

		fn := xlog.DebugCtx
		if lr.statusCode >= 500 {
			fn = xlog.ErrorCtx
		} else if lr.statusCode >= 400 {
			fn = xlog.WarnCtx
		} else if lr.statusCode >= 300 {
			fn = xlog.InfoCtx
		}

		fn(r.Context(), "request", slog.String("method", r.Method), slog.String("path", r.URL.Path), slog.Int("statusCode", lr.statusCode), slog.String("duration", duration.String()))
	})
}

// responseLogger 是 http.ResponseWriter 的一个封装
// 它允许我们捕获状态码
type responseLogger struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader 用来捕获状态码
func (rl *responseLogger) WriteHeader(code int) {
	rl.statusCode = code
	rl.ResponseWriter.WriteHeader(code)
}

func (rl *responseLogger) Flush() {
	// 断言原始的 ResponseWriter 支持 Flusher 接口
	flusher, ok := rl.ResponseWriter.(http.Flusher)
	if !ok {
		panic("原始 ResponseWriter 不支持 http.Flusher 接口")
	}
	flusher.Flush()
}
