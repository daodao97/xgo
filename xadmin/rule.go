package xadmin

import (
	"embed"
	_ "embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"

	"github.com/daodao97/xgo/xdb"
	"github.com/daodao97/xgo/xlog"
	"github.com/daodao97/xgo/xtype"
)

type Rule struct {
	Schema     string                                          `json:"schema"`
	NewModel   func(r *http.Request) xdb.Model                 `json:"new_model"`
	ListRule   func(r *http.Request) []xdb.Option              `json:"list_rule"`
	ViewRule   func(r *http.Request) []xdb.Option              `json:"view_rule"`
	UpdateRule func(r *http.Request) []xdb.Option              `json:"update_rule"`
	DeleteRule func(r *http.Request) []xdb.Option              `json:"delete_rule"`
	AfterList  func(r *http.Request, list []xdb.Row) []xdb.Row `json:"after_list"`
}

func (r Rule) GetSchema() any {
	var _any any
	if err := json.Unmarshal([]byte(r.Schema), &_any); err != nil {
		xlog.Error("Failed to unmarshal schema", xlog.Err(err))
	}
	return _any
}

func (r Rule) DecodeSchema() Schema {
	var s Schema
	if err := json.Unmarshal([]byte(r.Schema), &s); err != nil {
		xlog.Error("Failed to unmarshal schema", xlog.Err(err))
	}
	return s
}

var Rules = map[string]Rule{}

func InitSchema(schema embed.FS) {
	err := fs.WalkDir(schema, "schema", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".json") {
			// 获取文件内容并创建规则
			content, err := fs.ReadFile(schema, path)
			if err != nil {
				return err
			}
			_content := xtype.JsonStrVarReplace(string(content), map[string]any{})
			// 使用文件名（无扩展名）作为规则键
			key := strings.TrimSuffix(strings.TrimPrefix(path, "schema/"), ".json")
			exists, ok := Rules[key]
			if ok {
				exists.Schema = _content
				Rules[key] = exists
			} else {
				Rules[key] = Rule{
					Schema: _content,
				}
			}
		}
		return nil
	})

	if err != nil {
		xlog.Error("Failed to walk through schema directory: %v", err)
	}
}

func RegAfterList(collection string, fn func(r *http.Request, list []xdb.Row) []xdb.Row) {
	rule, ok := Rules[collection]
	if ok {
		rule.AfterList = fn
		Rules[collection] = rule
	} else {
		Rules[collection] = Rule{
			AfterList: fn,
		}
	}
}
