package xjson

import (
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type Json struct {
	data string
	*Var
}

func New(data any) *Json {
	return &Json{data: ToString(data), Var: NewVar(data)}
}

func (j *Json) Get(path string) *Var {
	return NewVar(gjson.Get(j.data, path).Raw)
}

func (j *Json) Set(path string, value any) *Json {
	body, _ := sjson.Set(j.data, path, value)
	return &Json{data: body}
}

func (j *Json) String() string {
	return j.data
}
