package xapp

import (
	"log/slog"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"

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
	// 添加 decimal 类型验证支持
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterCustomTypeFunc(func(field reflect.Value) interface{} {
			if value, ok := field.Interface().(decimal.Decimal); ok {
				return value.String()
			}
			return nil
		}, decimal.Decimal{})
	}

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

func HanderFunc[Req any, Resp any](handler func(*gin.Context, Req) (*Resp, error)) func(*gin.Context) {
	return func(c *gin.Context) {
		var req Req
		// 首先设置默认值
		setDefaultValues(&req)

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
