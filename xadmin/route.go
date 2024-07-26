package xadmin

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/daodao97/xgo/xhttp"
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

		_, err := _jwt.ParseToken(r.Header.Get("X-Token"))
		if err != nil {
			xhttp.ResponseJson(w, map[string]interface{}{
				"code": 401,
				"msg":  "Unauthorized" + err.Error(),
			})
			return
		}

		// 验证逻辑
		handler.ServeHTTP(w, r)
	})
}
