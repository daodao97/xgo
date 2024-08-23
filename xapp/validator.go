package xapp

import (
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
