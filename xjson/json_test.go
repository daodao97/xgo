package xjson

import (
	"testing"
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
	t.Log(json.Get("name").String())
	t.Log(json.Set("name", "daodao2").String())
	t.Log(json.Get("metadata").String())
	t.Log(json.Get("metadata").JSON().String())
	t.Log(json.Get("metadata").JSON().Set("hello", "xjson").String())
}
