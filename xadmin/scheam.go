package xadmin

import (
	"encoding/json"
	"fmt"
)

type SelectOption struct {
	Value interface{} `json:"value"`
	Label string      `json:"label"`
}

type Filter struct {
	Label    string         `json:"label"`    // 搜索项的名称, 无时显示 Field 字段的值
	Field    string         `json:"field"`    // 搜索项的字段名, 用于构造 db.Select 的条件, 必须
	Operator string         `json:"operator"` // 搜索项的操作符, 如 >, <, in 等, 默认 =
	Type     string         `json:"type"`
	Options  []SelectOption `json:"options"`
	Form     bool           `json:"form,omitempty"`
}

func (f *Filter) UnmarshalJSON(data []byte) error {
	f.Form = true
	type Alias Filter
	tmp := (*Alias)(f)

	return json.Unmarshal(data, tmp)
}

type Header struct {
	Label   string            `json:"label"` // 表头
	Field   string            `json:"field"` // 字段名
	Render  string            `json:"render,omitempty"`
	Type    string            `json:"type,omitempty"`
	State   map[string]string `json:"state,omitempty"`
	Options []SelectOption    `json:"options,omitempty"`
	Fake    bool              `json:"fake,omitempty"`
	HasMany string            `json:"hasMany,omitempty"`
}

func (h *Header) UnmarshalJSON(data []byte) error {
	h.Fake = false
	type Alias Header
	tmp := (*Alias)(h)

	return json.Unmarshal(data, tmp)
}

type Button struct {
	Text      string                 `json:"text"`
	Target    string                 `json:"target"`
	Tips      string                 `json:"tips,omitempty"`
	Type      string                 `json:"type"`
	Props     map[string]interface{} `json:"props,omitempty"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
	SubButton []Button               `json:"subButton,omitempty"`
}

type FormItems struct {
	Type     string `json:"type,omitempty"`
	Label    string `json:"label,omitempty"`
	Field    string `json:"field,omitempty"`
	Form     bool   `json:"form,omitempty"`
	Validate string `json:"validate,omitempty"`
}

func (fi *FormItems) UnmarshalJSON(data []byte) error {
	fi.Form = true
	type Alias FormItems
	tmp := (*Alias)(fi)

	return json.Unmarshal(data, tmp)
}

type Orderby struct {
	Field string `json:"field"`
	Mod   string `json:"mod"`
}

type Tab struct {
	Value    interface{}
	Label    string
	Field    string
	Operator string
}

type baseWhere struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
	Logic    string      `json:"logic"`
}

type Schema struct {
	// 列表相关
	ExportCurrentPageAble bool        `json:"exportCurrentPageAble,omitempty"` // 导出当前列表
	Tabs                  []Tab       `json:"tabs,omitempty"`                  // 列表的分tab
	OrderBy               *Orderby    `json:"orderBy,omitempty"`               // 列表数据的默认排序, 会被用户请求中 sort_field=xx&sort_type=desc|asc 同名field重载
	OrderByMulti          []Orderby   `json:"OrderByMulti,omitempty"`          // 列表数据的默认排序, 会被用户请求中 sort_field=xx&sort_type=desc|asc 同名field重载
	GroupBy               string      `json:"groupBy,omitempty"`               // 列表数据的默认分支, 会被用户请求中 group_by=xxx 重载
	BaseWhere             []baseWhere `json:"baseWhere,omitempty"`             // 列表数据的基础过滤项
	Filter                []Filter    `json:"filter,omitempty"`                // 列表筛选表单
	Headers               []Header    `json:"headers,omitempty"`               // 列表表头定义
	NormalButton          []Button    `json:"normalButton,omitempty"`          // 列表页的普通操作按钮
	RowButton             []Button    `json:"rowButton,omitempty"`             // 列表每行的操作按钮
	BatchButton           []Button    `json:"batchButton,omitempty"`           // 列表的批量操作按钮
	HasOne                []string    `json:"hasOne,omitempty"`                // [pool.]db.table:[local_key->]foreign_key,other_key
	ListApi               string      `json:"listApi,omitempty"`               // 列表数据拉去接口
	// 表单相关
	FormItems []FormItems `json:"formItems,omitempty"` // 表单的定义
	SaveApi   string      `json:"saveApi,omitempty"`
	GetApi    string      `json:"getApi,omitempty"`
	// 其他
	Maintenance bool `json:"maintenance,omitempty"` // 功能维护中
}

func (s *Schema) UnmarshalJSON(data []byte) error {
	s.Tabs = []Tab{}
	s.Filter = []Filter{}
	s.Headers = []Header{}
	s.NormalButton = []Button{}
	s.RowButton = []Button{}
	s.BatchButton = []Button{}
	s.FormItems = []FormItems{}
	type Alias Schema
	tmp := (*Alias)(s)

	return json.Unmarshal(data, tmp)
}

type TreeOptions struct {
	Value    string        `json:"value"`
	Label    string        `json:"label"`
	ModuleId int           `json:"-"`
	Children []TreeOptions `json:"children,omitempty"`
}

func PageSchemaResourceList(id int, pageSchema *Schema) []TreeOptions { //nolint:gocognit
	var options []TreeOptions

	var formItemOptions []TreeOptions
	if len(pageSchema.FormItems) > 0 {
		for _, v := range pageSchema.FormItems {
			formItemOptions = append(formItemOptions, TreeOptions{Value: fmt.Sprintf("%d.formItems.%s", id, v.Field), Label: v.Label})
		}
	}
	if len(formItemOptions) > 0 {
		options = append(options, TreeOptions{
			Value:    fmt.Sprintf("%d.%s", id, "formItems"),
			Label:    "表单项",
			Children: formItemOptions,
		})
	}

	if pageSchema.SaveApi != "" {
		options = append(options, TreeOptions{
			Value: fmt.Sprintf("%d.%s", id, "saveApi"),
			Label: "保存表单",
		})
	}

	var tableFilterOptions []TreeOptions
	if len(pageSchema.Filter) > 0 {
		for _, v := range pageSchema.Filter {
			tableFilterOptions = append(tableFilterOptions, TreeOptions{Value: fmt.Sprintf("%d.filter.%s", id, v.Field), Label: v.Label})
		}
	}
	if len(tableFilterOptions) > 0 {
		options = append(options, TreeOptions{
			Value:    fmt.Sprintf("%d.%s", id, "filter"),
			Label:    "筛选条件",
			Children: tableFilterOptions,
		})
	}
	var tableHeaderOptions []TreeOptions
	if len(pageSchema.Headers) > 0 {
		for _, v := range pageSchema.Headers {
			tableHeaderOptions = append(tableHeaderOptions, TreeOptions{Value: fmt.Sprintf("%d.headers.%s", id, v.Field), Label: v.Label})
		}
	}
	if len(tableHeaderOptions) > 0 {
		options = append(options, TreeOptions{
			Value:    fmt.Sprintf("%d.%s", id, "headers"),
			Label:    "列表项",
			Children: tableHeaderOptions,
		})
	}

	var tableNormalButton []TreeOptions
	if len(pageSchema.NormalButton) > 0 {
		for _, v := range pageSchema.NormalButton {
			text := v.Text
			if text == "" && v.Tips != "" {
				text = v.Tips
			}
			tableNormalButton = append(tableNormalButton, TreeOptions{Value: fmt.Sprintf("%d.normalButton.%s", id, text), Label: text})
		}
	}
	if len(tableNormalButton) > 0 {
		options = append(options, TreeOptions{
			Value:    fmt.Sprintf("%d.%s", id, "normalButton"),
			Label:    "列表按钮",
			Children: tableNormalButton,
		})
	}

	var tableBatchButton []TreeOptions
	if len(pageSchema.BatchButton) > 0 {
		for _, v := range pageSchema.BatchButton {
			text := v.Text
			if text == "" && v.Tips != "" {
				text = v.Tips
			}
			tableBatchButton = append(tableBatchButton, TreeOptions{Value: fmt.Sprintf("%d.batchButton.%s", id, text), Label: text})
		}
	}
	if len(tableBatchButton) > 0 {
		options = append(options, TreeOptions{
			Value:    fmt.Sprintf("%d.%s", id, "batchButton"),
			Label:    "批量按钮",
			Children: tableBatchButton,
		})
	}

	var tableRowButton []TreeOptions
	if len(pageSchema.RowButton) > 0 {
		for _, v := range pageSchema.RowButton {
			text := v.Text
			if text == "" && v.Tips != "" {
				text = v.Tips
			}
			tableRowButton = append(tableRowButton, TreeOptions{Value: fmt.Sprintf("%d.rowButton.%s", id, text), Label: text})
		}
	}
	if len(tableRowButton) > 0 {
		options = append(options, TreeOptions{
			Value:    fmt.Sprintf("%d.%s", id, "rowButton"),
			Label:    "行操作按钮",
			Children: tableRowButton,
		})
	}

	return options
}
