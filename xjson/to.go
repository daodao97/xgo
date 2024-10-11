package xjson

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/spf13/cast"
)

func jsonStringToObject(s string, v any) error {
	data := []byte(s)
	return json.Unmarshal(data, v)
}

func ToStringMapE(i any) (map[string]any, error) {
	var m = map[string]any{}

	switch v := i.(type) {
	case map[any]any:
		for k, val := range v {
			m[ToString(k)] = val
		}
		return m, nil
	case map[string]any:
		return v, nil
	case map[string]string:
		for k, val := range v {
			m[k] = val
		}
		return m, nil
	case string:
		err := jsonStringToObject(v, &m)
		return m, err
	case []byte:
		err := json.Unmarshal(v, &m)
		return m, err
	case io.Reader:
		dec := json.NewDecoder(v)
		err := dec.Decode(&m)
		return m, err
	default:
		// 尝试使用反射来处理结构体
		if reflect.TypeOf(i).Kind() == reflect.Struct {
			j, err := json.Marshal(i)
			if err != nil {
				return m, fmt.Errorf("unable to marshal struct to JSON: %v", err)
			}
			err = json.Unmarshal(j, &m)
			return m, err
		}
		return m, fmt.Errorf("unable to cast %#v of type %T to map[string]any", i, i)
	}
}

func ToSliceE(i any) ([]any, error) {
	var s []any

	switch v := i.(type) {
	case []any:
		return append(s, v...), nil
	case []map[string]any:
		for _, u := range v {
			s = append(s, u)
		}
		return s, nil
	case []string:
		for _, u := range v {
			s = append(s, u)
		}
		return s, nil
	case []int:
		for _, u := range v {
			s = append(s, u)
		}
		return s, nil
	case []float64:
		for _, u := range v {
			s = append(s, u)
		}
		return s, nil
	case string:
		// 尝试解析 JSON 字符串
		var jsonSlice []any
		err := json.Unmarshal([]byte(v), &jsonSlice)
		if err == nil {
			return jsonSlice, nil
		}
		// 如果不是有效的 JSON，则将字符串作为单个元素的切片返回
		return []any{v}, nil
	case []byte:
		var jsonSlice []any
		err := json.Unmarshal(v, &jsonSlice)
		if err == nil {
			return jsonSlice, nil
		}
		// 如果不是有效的 JSON，则将字符串作为单个元素的切片返回
		return []any{v}, nil
	default:
		// 使用反射处理其他可能的切片类型
		rv := reflect.ValueOf(i)
		if rv.Kind() == reflect.Slice {
			for j := 0; j < rv.Len(); j++ {
				s = append(s, rv.Index(j).Interface())
			}
			return s, nil
		}
		// 如果不是切片类型，则将其作为单个元素的切片返回
		return []any{i}, nil
	}
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
	case map[string]any, []any:
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
