package xapp

import (
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

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

func translateError(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var errMsgs []string
		for _, e := range validationErrors {
			switch e.Tag() {
			case "required":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 是必填字段", e.Field()))
			case "min":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 的长度不能小于 %s", e.Field(), e.Param()))
			case "max":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 的长度不能大于 %s", e.Field(), e.Param()))
			default:
				errMsgs = append(errMsgs, fmt.Sprintf("%s 字段验证失败", e.Field()))
			}
		}
		return strings.Join(errMsgs, "; ")
	}
	return err.Error()
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

		resp, err := handler(c, req)
		if err != nil {
			c.JSON(200, gin.H{
				"code":    500,
				"message": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"code":    0,
			"message": "success",
			"data":    resp,
		})
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

func setDefaultValues(obj interface{}) {
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
