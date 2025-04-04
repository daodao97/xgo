package xjson

import (
	"encoding/json"

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
	val := gjson.Get(j.data, path)
	return NewVar(val)
}

func (j *Json) Set(path string, value any) *Json {
	body, _ := sjson.Set(j.data, path, value)
	return &Json{data: body}
}

func (j *Json) String() string {
	return j.data
}

func (j *Json) Unmarshal(v any) error {
	return json.Unmarshal([]byte(j.data), v)
}
