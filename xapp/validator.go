package xapp

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Validator interface {
	Validate() error
}

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("enum", validateEnum)
		v.RegisterValidation("datetime_range", validateDateTimeRange)
	}
}

func validateEnum(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	enumValues := strings.Split(fl.Param(), "|")
	for _, enum := range enumValues {
		if value == enum {
			return true
		}
	}
	return false
}

const timeFormat = "2006-01-02 15:04:05"

func validateDateTimeRange(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}

	parts := strings.Split(value, ",")
	if len(parts) != 2 {
		return false
	}

	startTime, err := time.Parse(timeFormat, strings.TrimSpace(parts[0]))
	if err != nil {
		return false
	}

	endTime, err := time.Parse(timeFormat, strings.TrimSpace(parts[1]))
	if err != nil {
		return false
	}

	return !endTime.Before(startTime)
}

// https://liuqh.icu/2021/05/30/go/gin/11-validate/
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
			case "oneof":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须是 %s 中的一个", e.Field(), e.Param()))
			case "email":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须是有效的电子邮件地址", e.Field()))
			case "len":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 的长度必须等于 %s", e.Field(), e.Param()))
			case "eq":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须等于 %s", e.Field(), e.Param()))
			case "ne":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 不能等于 %s", e.Field(), e.Param()))
			case "lt":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须小于 %s", e.Field(), e.Param()))
			case "lte":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须小于或等于 %s", e.Field(), e.Param()))
			case "gt":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须大于 %s", e.Field(), e.Param()))
			case "gte":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须大于或等于 %s", e.Field(), e.Param()))
			case "alpha":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 只能包含字母", e.Field()))
			case "alphanum":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 只能包含字母和数字", e.Field()))
			case "numeric":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须是数字", e.Field()))
			case "required_with":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 是必填字段", e.Field()))
			case "omitempty":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 字段可以为空", e.Field()))
			case "eqfield":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须等于 %s 字段", e.Field(), e.Param()))
			case "nefield":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 不能等于 %s 字段", e.Field(), e.Param()))
			case "gtfield":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须大于 %s 字段", e.Field(), e.Param()))
			case "gtefield":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须大于或等于 %s 字段", e.Field(), e.Param()))
			case "ltfield":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须小于 %s 字段", e.Field(), e.Param()))
			case "ltefield":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须小于或等于 %s 字段", e.Field(), e.Param()))
			case "eqcsfield":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须等于 %s 字段", e.Field(), e.Param()))
			case "necsfield":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 不能等于 %s 字段", e.Field(), e.Param()))
			case "gtcsfield":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须大于 %s 字段", e.Field(), e.Param()))
			case "gtecsfield":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须大于或等于 %s 字段", e.Field(), e.Param()))
			case "ltcsfield":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须小于 %s 字段", e.Field(), e.Param()))
			case "ltecsfield":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须小于或等于 %s 字段", e.Field(), e.Param()))
			case "structonly":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 只验证结构体，不验证任何结构体字段", e.Field()))
			case "nostructlevel":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 不运行任何结构级别的验证", e.Field()))
			case "dive":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 验证失败", e.Field()))
			case "required_with_all":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 是必填字段，因为 %s 都不为空", e.Field(), e.Param()))
			case "required_without":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 是必填字段，因为 %s 中有字段为空", e.Field(), e.Param()))
			case "required_without_all":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 是必填字段，因为 %s 都为空", e.Field(), e.Param()))
			case "isdefault":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须是默认值", e.Field()))
			case "containsfield":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须包含 %s 字段", e.Field(), e.Param()))
			case "excludesfield":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 不能包含 %s 字段", e.Field(), e.Param()))
			case "unique":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须是唯一的", e.Field()))
			case "alphaunicode":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 只能包含 unicode 字符", e.Field()))
			case "alphanumunicode":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 只能包含 unicode 字母和数字", e.Field()))
			case "hexadecimal":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须是有效的十六进制", e.Field()))
			case "hexcolor":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须是有效的十六进制颜色", e.Field()))
			case "lowercase":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 只能包含小写字符", e.Field()))
			case "uppercase":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 只能包含大写字符", e.Field()))
			case "json":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须是有效的 JSON", e.Field()))
			case "file":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须是有效的文件路径", e.Field()))
			case "url":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须是有效的 URL", e.Field()))
			case "uri":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须是有效的 URI", e.Field()))
			case "base64":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须是有效的 base64 值", e.Field()))
			case "contains":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须包含 %s", e.Field(), e.Param()))
			case "containsany":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须包含 %s 中的任何字符", e.Field(), e.Param()))
			case "containsrune":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须包含字符 %s", e.Field(), e.Param()))
			case "excludes":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 不能包含 %s", e.Field(), e.Param()))
			case "excludesall":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 不能包含 %s 中的任何字符", e.Field(), e.Param()))
			case "excludesrune":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 不能包含字符 %s", e.Field(), e.Param()))
			case "startswith":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须以 %s 开始", e.Field(), e.Param()))
			case "endswith":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须以 %s 结束", e.Field(), e.Param()))
			case "ip":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须是有效的 IP 地址", e.Field()))
			case "ipv4":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须是有效的 IPv4 地址", e.Field()))
			case "datetime":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须是有效的日期时间", e.Field()))
			case "datetime_range":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 格式必须是 'YYYY-MM-DD HH:mm:ss,YYYY-MM-DD HH:mm:ss' 且结束时间必须大于开始时间", e.Field()))
			default:
				errMsgs = append(errMsgs, fmt.Sprintf("%s 字段验证失败", e.Field()))
			}
		}
		return strings.Join(errMsgs, "; ")
	}
	return err.Error()
}
