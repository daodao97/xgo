package xdb

import (
	"github.com/daodao97/xgo/xdb/interval/hook"
)

type HookData interface {
	Input(row map[string]interface{}, fieldValue interface{}) (interface{}, error)
	Output(row map[string]interface{}, fieldValue interface{}) (interface{}, error)
}

type Hook = func() (string, HookData)

func Json(field string) Hook {
	return func() (string, HookData) {
		return field, &hook.Json{}
	}
}

func Array(field string) Hook {
	return func() (string, HookData) {
		return field, &hook.Array{}
	}
}

func CommaInt(field string) Hook {
	return func() (string, HookData) {
		return field, &hook.CommaSeparatedInt{}
	}
}
