package hook

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cast"

	"github.com/daodao97/xgo/xdb/interval/util"
)

type Json struct{}

func (Json) Input(row map[string]any, fieldValue any) (any, error) {
	if fieldValue == nil {
		return nil, nil
	}
	bt, err := json.MarshalIndent(fieldValue, "", "  ")
	if err != nil {
		return nil, err
	}
	return string(bt), err
}

func (Json) Output(row map[string]any, fieldValue any) (any, error) {
	str := cast.ToString(fieldValue)
	if str == "" {
		return nil, nil
	}
	str, err := util.JsonStrRemoveComments(str)
	if err != nil {
		return nil, err
	}
	tmp1 := new([]any)
	err1 := json.Unmarshal([]byte(str), tmp1)
	if err1 == nil {
		return tmp1, nil
	}
	tmp2 := new(map[string]any)
	err2 := json.Unmarshal([]byte(str), tmp2)
	if err2 == nil {
		return tmp2, nil
	}

	return nil, fmt.Errorf("Hook.Json.Output err %v %v", err1, err2)
}

/** Array columnHook **/

type Array struct {
	Json
}

func (Array) Output(row map[string]any, fieldValue any) (any, error) {
	str := cast.ToString(fieldValue)
	tmp1 := new([]any)
	if str == "" {
		return tmp1, nil
	}
	str, err := util.JsonStrRemoveComments(str)
	if err != nil {
		return nil, err
	}
	err1 := json.Unmarshal([]byte(str), tmp1)
	if err1 == nil {
		return tmp1, nil
	}

	return nil, fmt.Errorf("Hook.Array.Output err %v", err1)
}

/** Object columnHook **/

type Object struct {
	Json
}

func (Object) Output(row map[string]any, fieldValue any) (any, error) {
	str := cast.ToString(fieldValue)
	tmp2 := new(map[string]any)
	if str == "" {
		return tmp2, nil
	}
	str, err := util.JsonStrRemoveComments(str)
	if err != nil {
		return nil, err
	}
	err2 := json.Unmarshal([]byte(str), tmp2)
	if err2 == nil {
		return tmp2, nil
	}

	return nil, fmt.Errorf("Hook.Object.Output err %v", err2)
}
