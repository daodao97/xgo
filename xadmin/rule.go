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

type Crud struct {
	Schema       string
	NewModel     func(r *http.Request) xdb.Model                         // 创建自定义 model
	BeforeGet    func(r *http.Request, opt []xdb.Option) []xdb.Option    // 在获取单个之前执行, 修改筛选条件
	BeforeCreate func(r *http.Request, createData xdb.Record) xdb.Record // 在创建之前执行, 修改创建数据
	BeforeUpdate func(r *http.Request, updateData xdb.Record) xdb.Record // 在更新之前执行, 修改更新数据
	BeforeDelete func(r *http.Request, opt []xdb.Option) []xdb.Option    // 在删除之前执行, 修改筛选条件
	BeforeList   func(r *http.Request, opt []xdb.Option) []xdb.Option    // 在列表之前执行, 修改筛选条件
	AfterList    func(r *http.Request, list []xdb.Record) []xdb.Record   // 在列表之后执行, 修改列表数据
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

func RegAfterList(collection string, fn func(r *http.Request, list []xdb.Record) []xdb.Record) {
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

func RegBeforeList(collection string, fn func(r *http.Request, opt []xdb.Option) []xdb.Option) {
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

func RegBeforeGet(collection string, fn func(r *http.Request, opt []xdb.Option) []xdb.Option) {
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

func RegBeforeCreate(collection string, fn func(r *http.Request, createData xdb.Record) xdb.Record) {
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

func RegBeforeUpdate(collection string, fn func(r *http.Request, updateData xdb.Record) xdb.Record) {
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

func RegBeforeDelete(collection string, fn func(r *http.Request, opt []xdb.Option) []xdb.Option) {
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
