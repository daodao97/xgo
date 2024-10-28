package xapp

import (
	"log/slog"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/daodao97/xgo/triceid"
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
	r.Use(gin.Recovery())
	r.Use(triceid.TraceId())
	r.Use(func(c *gin.Context) {
		// 开始时间
		start := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		end := time.Now()
		latency := end.Sub(start)

		// 使用 slog 记录结构化日志
		xlog.DebugC(c, "http request",
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

func HanderFunc[Req any, Resp any](handler func(*gin.Context, Req) (*Resp, error)) func(*gin.Context) {
	return func(c *gin.Context) {
		var req Req
		if err := c.ShouldBind(&req); err != nil {
			if err := c.ShouldBindQuery(&req); err != nil {
				setDefaultValues(&req)
				if err := c.ShouldBindQuery(&req); err != nil {
					c.JSON(200, gin.H{
						"code":    400,
						"message": translateError(err),
					})
					return
				}
			}
		}

		setDefaultValues(&req)

		validator, ok := any(&req).(Validator)
		if ok {
			// 如果转换成功，调用 Validate 方法
			if err := validator.Validate(); err != nil {
				// 处理验证错误
				c.JSON(200, gin.H{
					"code":    400,
					"message": translateError(err),
				})
				return
			}
			// 验证通过，继续处理
		}

		resp, err := handler(c, req)
		if err != nil {
			c.JSON(200, gin.H{
				"code":    500,
				"message": err.Error(),
			})
			return
		}

		body := gin.H{
			"code":    0,
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
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
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
