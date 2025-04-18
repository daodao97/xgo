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

type BeforeGetHook func(r *http.Request, opt []xdb.Option) []xdb.Option
type AfterGetHook func(r *http.Request, record xdb.Record) xdb.Record
type BeforeCreateHook func(r *http.Request, createData xdb.Record) (xdb.Record, error)
type AfterCreateHook func(r *http.Request, id int64, record xdb.Record)
type BeforeUpdateHook func(r *http.Request, updateData xdb.Record) (xdb.Record, error)
type AfterUpdateHook func(r *http.Request, id int64, record xdb.Record)
type BeforeDeleteHook func(r *http.Request, opt []xdb.Option) []xdb.Option
type AfterDeleteHook func(r *http.Request, id int64)
type BeforeListHook func(r *http.Request, opt []xdb.Option) []xdb.Option
type AfterListHook func(r *http.Request, list []xdb.Record) []xdb.Record
type SchemaHook func(r *http.Request, schema string) string
type Crud struct {
	Schema       string
	NewModel     func(r *http.Request) xdb.Model // 创建自定义 model
	BeforeGet    BeforeGetHook
	AfterGet     AfterGetHook
	BeforeCreate BeforeCreateHook
	AfterCreate  AfterCreateHook
	BeforeUpdate BeforeUpdateHook
	AfterUpdate  AfterUpdateHook
	BeforeDelete BeforeDeleteHook
	AfterDelete  AfterDeleteHook
	BeforeList   BeforeListHook
	AfterList    AfterListHook
	SchemaHook   SchemaHook
}

func (r Crud) GetSchema() any {
	var _any any
	if err := json.Unmarshal([]byte(r.Schema), &_any); err != nil {
		xlog.Error("Failed to unmarshal schema", xlog.Err(err))
	}
	return _any
}

func (r Crud) DecodeSchema() Schema {
	var s Schema
	if err := json.Unmarshal([]byte(r.Schema), &s); err != nil {
		xlog.Error("Failed to unmarshal schema", xlog.Err(err))
	}
	return s
}

var Cruds = map[string]Crud{}

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
			exists, ok := Cruds[key]
			if ok {
				exists.Schema = _content
				Cruds[key] = exists
			} else {
				Cruds[key] = Crud{
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

func RegSchemaHook(collection string, fn SchemaHook) {
	rule, ok := Cruds[collection]
	if ok {
		rule.SchemaHook = fn
		Cruds[collection] = rule
	} else {
		Cruds[collection] = Crud{
			SchemaHook: fn,
		}
	}
}

func RegAfterList(collection string, fn AfterListHook) {
	rule, ok := Cruds[collection]
	if ok {
		rule.AfterList = fn
		Cruds[collection] = rule
	} else {
		Cruds[collection] = Crud{
			AfterList: fn,
		}
	}
}

func RegBeforeList(collection string, fn BeforeListHook) {
	rule, ok := Cruds[collection]
	if ok {
		rule.BeforeList = fn
		Cruds[collection] = rule
	} else {
		Cruds[collection] = Crud{
			BeforeList: fn,
		}
	}
}

func RegBeforeGet(collection string, fn BeforeGetHook) {
	rule, ok := Cruds[collection]
	if ok {
		rule.BeforeGet = fn
		Cruds[collection] = rule
	} else {
		Cruds[collection] = Crud{
			BeforeGet: fn,
		}
	}
}

func RegAfterGet(collection string, fn AfterGetHook) {
	rule, ok := Cruds[collection]
	if ok {
		rule.AfterGet = fn
		Cruds[collection] = rule
	} else {
		Cruds[collection] = Crud{
			AfterGet: fn,
		}
	}
}

func RegBeforeCreate(collection string, fn BeforeCreateHook) {
	rule, ok := Cruds[collection]
	if ok {
		rule.BeforeCreate = fn
		Cruds[collection] = rule
	} else {
		Cruds[collection] = Crud{
			BeforeCreate: fn,
		}
	}
}

func RegAfterCreate(collection string, fn AfterCreateHook) {
	rule, ok := Cruds[collection]
	if ok {
		rule.AfterCreate = fn
		Cruds[collection] = rule
	} else {
		Cruds[collection] = Crud{
			AfterCreate: fn,
		}
	}
}

func RegBeforeUpdate(collection string, fn BeforeUpdateHook) {
	rule, ok := Cruds[collection]
	if ok {
		rule.BeforeUpdate = fn
		Cruds[collection] = rule
	} else {
		Cruds[collection] = Crud{
			BeforeUpdate: fn,
		}
	}
}

func RegAfterUpdate(collection string, fn AfterUpdateHook) {
	rule, ok := Cruds[collection]
	if ok {
		rule.AfterUpdate = fn
		Cruds[collection] = rule
	} else {
		Cruds[collection] = Crud{
			AfterUpdate: fn,
		}
	}
}

func RegBeforeDelete(collection string, fn BeforeDeleteHook) {
	rule, ok := Cruds[collection]
	if ok {
		rule.BeforeDelete = fn
		Cruds[collection] = rule
	} else {
		Cruds[collection] = Crud{
			BeforeDelete: fn,
		}
	}
}

func RegAfterDelete(collection string, fn AfterDeleteHook) {
	rule, ok := Cruds[collection]
	if ok {
		rule.AfterDelete = fn
		Cruds[collection] = rule
	} else {
		Cruds[collection] = Crud{
			AfterDelete: fn,
		}
	}
}

func RegNewModel(collection string, fn func(r *http.Request) xdb.Model) {
	rule, ok := Cruds[collection]
	if ok {
		rule.NewModel = fn
		Cruds[collection] = rule
	} else {
		Cruds[collection] = Crud{
			NewModel: fn,
		}
	}
}
