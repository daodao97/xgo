package xapp

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/daodao97/xgo/utils"
	"github.com/daodao97/xgo/xlog"
)

type SlogWriter struct {
	logger *slog.Logger
}

func (w SlogWriter) Write(p []byte) (n int, err error) {
	w.logger.Info(string(p))
	return len(p), nil
}

func NewGin() *gin.Engine {
	logger := SlogWriter{logger: xlog.GetLogger()}
	gin.DefaultWriter = logger
	gin.DefaultErrorWriter = logger
	if !utils.IsGoRun() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(func(c *gin.Context) {
		// 开始时间
		start := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		end := time.Now()
		latency := end.Sub(start)

		// 使用 slog 记录结构化日志
		xlog.Info("HTTP Request",
			slog.String("client_ip", c.ClientIP()),
			slog.String("time", end.Format(time.DateTime)),
			slog.Int("status_code", c.Writer.Status()),
			slog.String("latency", latency.String()),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("error_message", c.Errors.ByType(gin.ErrorTypePrivate).String()),
		)
	})
	return r
}
