package xjson

import "github.com/spf13/cast"

func NewVar(data any) *Var {
	return &Var{data: data}
}

type Var struct {
	data any
}

func (v *Var) String() string {
	return ToString(v.data)
}

func (v *Var) Int() int {
	return cast.ToInt(v.data)
}

func (v *Var) Int64() int64 {
	return cast.ToInt64(v.data)
}

func (v *Var) Int32() int32 {
	return cast.ToInt32(v.data)
}

func (v *Var) Float() float64 {
	return cast.ToFloat64(v.data)
}

func (v *Var) Bool() bool {
	return cast.ToBool(v.data)
}

func (v *Var) JSON() *Json {
	return New(v.data)
}

func (v *Var) Array() []any {
	a, _ := ToSliceE(v.data)
	return a
}

func (v *Var) Map() map[string]any {
	m, _ := ToStringMapE(v.data)
	return m
}

func (v *Var) Slice() []any {
	a, _ := ToSliceE(v.data)
	return a
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
