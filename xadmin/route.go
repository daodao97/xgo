package xadmin

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed ui
var defaultUI embed.FS

// customUI 将允许替换的 UI，使用 fs.FS 以兼容任何文件系统实现
var customUI fs.FS

// SetUI 允许外部设置自定义 UI
func SetUI(ui fs.FS) {
	customUI = ui
}

var adminPath = "/_"

func SetAdminPath(path string) {
	adminPath = path
}

// httpHandlerToGin 将 http.HandlerFunc 转换为 gin.HandlerFunc
func httpHandlerToGin(f func(w http.ResponseWriter, r *http.Request)) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建一个 ResponseWriter 的包装器
		writer := &responseWriterWrapper{ResponseWriter: c.Writer, statusCode: http.StatusOK}

		// 调用原始的 http.HandlerFunc
		f(writer, c.Request)

		// 如果状态码已经被设置，则使用 Gin 的方法设置状态码
		c.Status(writer.statusCode)
	}
}

// responseWriterWrapper 包装 gin.ResponseWriter 以捕获状态码
type responseWriterWrapper struct {
	gin.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
