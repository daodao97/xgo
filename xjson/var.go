package xjson

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func NewVar(data any) *Var {
	if val, ok := data.(gjson.Result); ok {
		return &Var{Result: &val}
	}
	val := gjson.Parse(ToString(data))
	return &Var{Result: &val}
}

type Var struct {
	*gjson.Result
}

func (v *Var) IsNil() bool {
	return v.Result.Type == gjson.Null
}

func (v *Var) Int64() int64 {
	return v.Result.Int()
}

func (v *Var) Int32() int32 {
	return int32(v.Result.Int())
}

func (v *Var) JSON() *Json {
	return New(v.Result.Raw)
}

func (v *Var) Array() []any {
	arr := v.Result.Array()
	slice := make([]any, len(arr))
	for i, v := range arr {
		slice[i] = v.Raw
	}
	return slice
}

func (v *Var) Map() map[string]any {
	m := v.Result.Map()
	m2 := make(map[string]any)
	for k, v := range m {
		m2[k] = v.Raw
	}
	return m2
}

func (v *Var) MapString() map[string]string {
	m := v.Result.Map()
	m2 := make(map[string]string)
	for k, v := range m {
		m2[k] = ToString(v.Raw)
	}
	return m2
}

func (v *Var) Slice() []any {
	arr := v.Result.Array()
	slice := make([]any, len(arr))
	for i, v := range arr {
		slice[i] = v.Raw
	}
	return slice
}

func (v *Var) ArrayJson() []*Json {
	arr := v.Array()
	jsonArr := make([]*Json, len(arr))
	for i, v := range arr {
		jsonArr[i] = New(v)
	}
	return jsonArr
}

func (v *Var) MapJson() map[string]*Json {
	arr := v.Map()
	jsonArr := make(map[string]*Json)
	for i, v := range arr {
		jsonArr[i] = New(v)
	}
	return jsonArr
}

func (v *Var) TimeByFormat(format string) time.Time {
	res, _ := time.Parse(format, v.String())
	return res
}

func (v *Var) Decimal() decimal.Decimal {
	d, _ := decimal.NewFromString(v.String())
	return d
}
