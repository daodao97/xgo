package xadmin

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/daodao97/xgo/xhttp"
	"github.com/daodao97/xgo/xjwt"
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

func Route(r *mux.Router) *mux.Router {
	_ui := fs.FS(defaultUI)
	if customUI != nil {
		_ui = customUI
	}

	contentStatic, err := fs.Sub(_ui, "ui")
	if err != nil {
		panic(err)
	}

	// 创建文件服务器
	fileServer := http.FileServer(http.FS(contentStatic))
	r.PathPrefix(adminPath + "/").Handler(http.StripPrefix(adminPath+"/", fileServer)).Methods(http.MethodGet)

	api := r.PathPrefix(fmt.Sprintf("%sapi", adminPath)).Subrouter()
	api.Use(auth)
	UserRoute(api)
	api.HandleFunc("/schema/{table_name}", PageSchema)
	api.HandleFunc("/{table_name}/create", Create).Methods(http.MethodPost)
	api.HandleFunc("/{table_name}/list", List).Methods(http.MethodGet)
	api.HandleFunc("/{table_name}/get/{id}", Read).Methods(http.MethodGet)
	api.HandleFunc("/{table_name}/update/{id}", Update).Methods(http.MethodPost)
	api.HandleFunc("/{table_name}/del", Delete).Methods(http.MethodDelete)
	api.HandleFunc("/{table_name}/options", Options).Methods(http.MethodGet)

	return api

}

func auth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == fmt.Sprintf("%sapi/user/login", adminPath) {
			handler.ServeHTTP(w, r)
			return
		}

		if r.Header.Get("X-Token") == "" {
			xhttp.ResponseJson(w, map[string]any{
				"code": 401,
			})
			return
		}

		_, err := xjwt.VerifyHMacToken(r.Header.Get("X-Token"), _jwtConf.Secret)
		if err != nil {
			xhttp.ResponseJson(w, map[string]any{
				"code": 401,
				"msg":  "Unauthorized" + err.Error(),
			})
			return
		}

		handler.ServeHTTP(w, r)
	})
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
