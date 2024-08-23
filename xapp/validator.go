package xapp

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("enum", validateEnum)
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
			case "enum":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须是预定义的枚举值之一", e.Field()))
			default:
				errMsgs = append(errMsgs, fmt.Sprintf("%s 字段验证失败", e.Field()))
			}
		}
		return strings.Join(errMsgs, "; ")
	}
	return err.Error()
}
