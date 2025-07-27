package xapp

import (
	"bytes"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"

	"github.com/daodao97/xgo/utils"
	"github.com/daodao97/xgo/xlog"
	"github.com/daodao97/xgo/xtrace"
	"github.com/daodao97/xgo/xutil"
)

type SlogWriter struct {
}

func (w SlogWriter) Write(p []byte) (n int, err error) {
	if strings.Contains(string(p), "[Recovery]") {
		return 0, nil
	}
	xlog.Info("gin log", xlog.String("info", string(p)))
	return len(p), nil
}

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func NewGin() *gin.Engine {
	// 添加 decimal 类型验证支持
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterCustomTypeFunc(func(field reflect.Value) interface{} {
			if value, ok := field.Interface().(decimal.Decimal); ok {
				return value.String()
			}
			return nil
		}, decimal.Decimal{})
	}

	logger := SlogWriter{}
	gin.DefaultWriter = logger
	gin.DefaultErrorWriter = logger
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		xlog.Debug("route", xlog.String("method", httpMethod), xlog.String("path", absolutePath), xlog.String("handler", handlerName))
	}
	if IsProd() {
		gin.SetMode(gin.ReleaseMode)
	} else if IsTest() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	//gin.DebugPrintFunc = func(format string, values ...interface{}) {}
	if !utils.IsGoRun() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	// r.Use(gin.Recovery())
	r.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		xlog.ErrorC(c, "panic recovered",
			xlog.Time("time", time.Now()),
			xlog.String("path", path),
			xlog.String("query", query),
			xlog.Any("error", err),
			// 可选：添加堆栈信息
			xlog.String("stack", string(xutil.Stack(3))),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "internal server error",
		})
	}))
	r.Use(xtrace.TraceId())
	r.Use(func(c *gin.Context) {
		// 检查是否为静态文件请求(.js, .css等)
		path := c.Request.URL.Path
		if isStaticFileRequest(path) {
			c.Next()
			return
		}

		// 开始时间
		start := time.Now()

		// 保存请求体
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// 包装响应写入器
		w := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = w

		// 处理请求
		c.Next()

		// 结束时间
		end := time.Now()
		duration := end.Sub(start)

		// 根据状态码选择日志级别
		logFunc := xlog.DebugC
		if c.Writer.Status() >= 400 {
			logFunc = xlog.WarnC
		}
		if c.Writer.Status() >= 500 {
			logFunc = xlog.ErrorC
		}

		args := []any{
			slog.String("client_ip", c.ClientIP()),
			slog.String("time", end.Format(time.DateTime)),
			slog.Int("status_code", c.Writer.Status()),
			slog.Duration("duration", duration),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
		}

		// 添加请求头信息
		headers := make(map[string]string)
		for k, v := range c.Request.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}
		args = append(args, slog.Any("headers", headers))

		// 检查是否为文件上传请求
		contentType := c.GetHeader("Content-Type")
		isMultipart := strings.HasPrefix(contentType, "multipart/form-data")

		// 如果不是文件上传，则记录请求体
		if !isMultipart && len(bodyBytes) > 0 {
			const maxBodyLength = 1024 * 10 // 10KB
			bodyStr := string(bodyBytes)
			if len(bodyStr) > maxBodyLength {
				bodyStr = bodyStr[:maxBodyLength] + "..."
			}
			args = append(args, slog.String("body", bodyStr))
		}

		// 获取响应的 Content-Type
		responseContentType := w.ResponseWriter.Header().Get("Content-Type")

		// 添加响应体
		if strings.HasPrefix(responseContentType, "application/json") && w.body.Len() > 0 {
			const maxRespLength = 1024 * 10 // 10KB
			respStr := w.body.String()
			if len(respStr) > maxRespLength {
				respStr = respStr[:maxRespLength] + "..."
			}
			args = append(args, slog.String("response", respStr))
		}

		logFunc(c, "http request", args...)
	})

	return r
}

func NewGinHttpServer(addr string, engine func() *gin.Engine) NewServer {
	return func() Server {
		return &HTTPServer{
			server: &http.Server{
				Addr:    addr,
				Handler: engine(),
			},
		}
	}
}

var SuccessCode = 0

func SetSuccessCode(code int) {
	SuccessCode = code
}

var MaxMultipartFormSize int64 = 32 << 20 // 32MB

func SetMaxMultipartFormSize(size int64) {
	MaxMultipartFormSize = size
}

type FileUploadRequest struct {
	Name  string                  `form:"name"`
	File  *multipart.FileHeader   `form:"file"`  // 单文件
	Files []*multipart.FileHeader `form:"files"` // 多文件
}

func HanderFunc[Req any, Resp any](handler func(*gin.Context, Req) (*Resp, error)) func(*gin.Context) {
	return func(c *gin.Context) {
		var req Req
		// 首先设置默认值
		setDefaultValues(&req)

		// 保存请求体
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// 检查是否为文件上传请求
		contentType := c.GetHeader("Content-Type")
		isMultipart := strings.HasPrefix(contentType, "multipart/form-data")

		// 如果是文件上传请求，先处理文件
		if isMultipart {
			if err := c.Request.ParseMultipartForm(MaxMultipartFormSize); err != nil {
				c.JSON(200, gin.H{
					"code":    400,
					"message": "文件上传失败: " + err.Error(),
				})
				return
			}
			// 对于文件上传请求，直接使用 ShouldBind
			if err := c.ShouldBind(&req); err != nil {
				c.JSON(200, gin.H{
					"code":    400,
					"message": translateError(err),
				})
				return
			}
		} else {
			c.ShouldBindUri(&req)
			c.ShouldBindQuery(&req)
			err3 := c.ShouldBind(&req)
			if err3 != nil {
				c.JSON(200, gin.H{
					"code":    400,
					"message": translateError(err3),
				})
				return
			}
		}

		if validator, ok := any(&req).(Validator); ok {
			if err := validator.Validate(); err != nil {
				c.JSON(200, gin.H{
					"code":    400,
					"message": translateError(err),
				})
				return
			}
		}

		resp, err := handler(c, req)
		if err != nil {
			c.JSON(500, gin.H{
				"code":    500,
				"message": err.Error(),
			})
			return
		}

		isSSE := c.Writer.Header().Get("Content-Type") == "text/event-stream"
		if isSSE {
			return
		}

		body := gin.H{
			"code":    SuccessCode,
			"message": "success",
		}

		if resp != nil {
			body["data"] = resp
		}

		c.JSON(200, body)
	}
}

type defaultSetter struct {
	fieldIndex int
	setFunc    func(reflect.Value, string)
}

var defaultSettersCache sync.Map // 用于缓存每种类型的默认值设置器

func getDefaultSetters(t reflect.Type) []defaultSetter {
	if setters, ok := defaultSettersCache.Load(t); ok {
		return setters.([]defaultSetter)
	}

	var setters []defaultSetter
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if defaultTag := field.Tag.Get("default"); defaultTag != "" {
			setter := defaultSetter{fieldIndex: i}
			switch field.Type.Kind() {
			case reflect.String:
				setter.setFunc = func(v reflect.Value, tag string) { v.SetString(tag) }
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				setter.setFunc = func(v reflect.Value, tag string) {
					if intValue, err := strconv.ParseInt(tag, 10, 64); err == nil {
						v.SetInt(intValue)
					}
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				setter.setFunc = func(v reflect.Value, tag string) {
					if uintValue, err := strconv.ParseUint(tag, 10, 64); err == nil {
						v.SetUint(uintValue)
					}
				}
			case reflect.Float32, reflect.Float64:
				setter.setFunc = func(v reflect.Value, tag string) {
					if floatValue, err := strconv.ParseFloat(tag, 64); err == nil {
						v.SetFloat(floatValue)
					}
				}
			case reflect.Bool:
				setter.setFunc = func(v reflect.Value, tag string) {
					if boolValue, err := strconv.ParseBool(tag); err == nil {
						v.SetBool(boolValue)
					}
				}
			}
			setters = append(setters, setter)
		}
	}

	defaultSettersCache.Store(t, setters)
	return setters
}

func setDefaultValues(obj any) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr {
		return
	}
	v = v.Elem()

	// 检查是否为结构体类型
	if v.Kind() != reflect.Struct {
		return
	}

	t := v.Type()
	setters := getDefaultSetters(t)
	for _, setter := range setters {
		field := v.Field(setter.fieldIndex)
		if field.IsZero() {
			defaultTag := t.Field(setter.fieldIndex).Tag.Get("default")
			setter.setFunc(field, defaultTag)
		}
	}
}

// 判断是否为静态文件请求
func isStaticFileRequest(path string) bool {
	staticFileExtensions := []string{".js", ".css", ".jpg", ".jpeg", ".png", ".gif", ".ico", ".svg", ".woff", ".woff2", ".ttf", ".eot"}
	for _, ext := range staticFileExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}
