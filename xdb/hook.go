package xdb

import (
	"github.com/daodao97/xgo/xdb/interval/hook"
)

type HookData interface {
	Input(row map[string]any, fieldValue any) (any, error)
	Output(row map[string]any, fieldValue any) (any, error)
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

func CommaString(field string) Hook {
	return func() (string, HookData) {
		return field, &hook.CommaSeparatedString{}
	}
}

func Time(field string, format string) Hook {
	return func() (string, HookData) {
		return field, &hook.Time{Format: format}
	}
}
