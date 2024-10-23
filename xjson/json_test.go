package xjson

import (
	"testing"

	"github.com/tidwall/gjson"
)

func TestJson(t *testing.T) {
	var data any
	data = map[string]any{"name": "daodao", "age": 18, "metadata": map[string]any{"name": "daodao"}}
	data = `{"name":"daodao","age":18,"metadata":{"name":"daodao"}}`
	data = struct {
		Name     string         `json:"name"`
		Age      int            `json:"age"`
		Metadata map[string]any `json:"metadata"`
	}{
		Name: "daodao",
		Age:  10,
		Metadata: map[string]any{
			"name": "daodao",
		},
	}

	json := New(data)
	t.Log(json.Map())
	t.Log(json.Get("name").String())
	t.Log(json.Set("name", "daodao2").String())
	t.Log(json.Get("metadata").String())
	t.Log(json.Get("metadata").JSON().String())
	t.Log(json.Get("metadata").JSON().Set("hello", "xjson").String())
}

var data = []byte(`{
    "string": "0009399372024102214115250-0000000005383248149",
    "int": 100000000,
	"array": [1,2,3],
	"map": {"a": 1, "b": 2, "c": {"d": 3}},
	"bool": true,
	"null": null
}`)

func TestJson2(t *testing.T) {
	json := New(data)
	t.Logf("string: %+v", json.Get("string").String())
	t.Logf("int: %+v", json.Get("int").Int())
	t.Logf("array: %+v", json.Get("array").Array())
	t.Logf("map: %+v", json.Get("map").Map())
	t.Logf("map.child: %+v", json.Get("map.a").Int())
	t.Logf("map.child2: %+v", json.Get("map.c"))
	t.Logf("bool: %+v", json.Get("bool").Bool())
	t.Logf("null: %+v", json.Get("null").IsNil())
}

func TestJson3(t *testing.T) {
	json := gjson.Parse(`"abc"`)
	t.Log(json.Type, json.Str)
}
