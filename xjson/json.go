package xjson

import (
	"encoding/json"
	"io"

	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type Json struct {
	data string
}

func New(data any) *Json {
	return &Json{data: ToString(data)}
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

func ToString(data any) string {
	var str string
	// fmt.Printf("数据类型: %T\n", data) // 添加此行来打印类型
	switch v := data.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	case io.Reader:
		b, err := io.ReadAll(v)
		if err == nil {
			str = string(b)
		}
	case map[string]interface{}, []interface{}:
		b, err := json.Marshal(v)
		if err == nil {
			str = string(b)
		}
	default:
		// 尝试将任何类型转换为 JSON 字符串
		b, err := json.Marshal(v)
		if err == nil {
			str = string(b)
		} else {
			// 如果无法转换为 JSON,则使用 cast.ToString
			str = cast.ToString(v)
		}
	}

	return str
}
