package xadmin

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/spf13/cast"

	"github.com/daodao97/xgo/xdb"
	"github.com/daodao97/xgo/xhttp"
)

func PageSchema(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	table := vars["table_name"]

	schema, ok := Cruds[table]
	if !ok {
		xhttp.ResponseJson(w, map[string]any{
			"code":    400,
			"message": "not support collection",
		})
		return
	}

	xhttp.ResponseJson(w, map[string]any{
		"code": 0,
		"msg":  "success",
		"data": schema.GetSchema(),
	})

}

func List(w http.ResponseWriter, r *http.Request) {
	table := xhttp.Vars(r, "table_name")

	schema, ok := Cruds[table]
	if !ok {
		xhttp.ResponseJson(w, map[string]any{
			"code":    400,
			"message": "not support collection",
		})
		return
	}

	ps := xhttp.QueryDefault(r, "_ps", "20")
	pn := xhttp.QueryDefault(r, "_pn", "1")

	m := xdb.New(table)
	if schema.NewModel != nil {
		m = schema.NewModel(r)
	}

	_schema := schema.DecodeSchema()
	var opt []xdb.Option
	for _, field := range _schema.Filter {
		value := r.URL.Query().Get(field.Field)
		if value != "" {
			switch field.Operator {
			case "like":
				opt = append(opt, xdb.Where(field.Field, "like", "%"+value+"%"))
			default:
				opt = append(opt, xdb.Where(field.Field, "=", value))
			}
		}
	}

	if schema.BeforeList != nil {
		opt = schema.BeforeList(r, opt)
	}

	count, err := m.Ctx(r.Context()).Count(opt...)
	if err != nil {
		xhttp.ResponseJson(w, map[string]any{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	if count == 0 {
		xhttp.ResponseJson(w, map[string]any{
			"code": 0,
			"msg":  "success",
			"data": map[string]any{
				"list": []any{},
				"page": map[string]any{
					"pn":    cast.ToInt(pn),
					"ps":    cast.ToInt(ps),
					"total": count,
				},
			},
		})
		return
	}

	var fields []string

	for _, field := range _schema.Headers {
		if field.Fake {
			continue
		}
		fields = append(fields, fmt.Sprintf("`%s`", field.Field))
	}

	opt = append(opt, xdb.Field(fields...))
	opt = append(opt, xdb.Pagination(cast.ToInt(pn), cast.ToInt(ps))...)

	orderBy := new(Orderby)
	if _schema.OrderBy != nil {
		orderBy = _schema.OrderBy
	}

	sortBy := xhttp.Query(r, "_sort_by")
	sortType := xhttp.QueryDefault(r, "_sort_type", "desc")
	if sortBy != "" {
		orderBy = &Orderby{
			Field: sortBy,
			Mod:   sortType,
		}
	}

	if orderBy != nil {
		switch orderBy.Mod {
		case "asc":
			opt = append(opt, xdb.OrderByAsc(orderBy.Field))
		case "desc":
			opt = append(opt, xdb.OrderByDesc(orderBy.Field))
		}
	}

	rows, err := m.Ctx(r.Context()).Selects(opt...)
	if err != nil {
		xhttp.ResponseJson(w, map[string]any{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	list := make([]xdb.Record, 0)
	if len(rows) > 0 {
		list = rows
	}

	if schema.AfterList != nil {
		list = schema.AfterList(r, list)
	}

	xhttp.ResponseJson(w, map[string]any{
		"code": 0,
		"msg":  "success",
		"data": map[string]any{
			"list": list,
			"page": map[string]any{
				"pn":    cast.ToInt(pn),
				"ps":    cast.ToInt(ps),
				"total": count,
			},
		},
	})
}

func Create(w http.ResponseWriter, r *http.Request) {
	table := xhttp.Vars(r, "table_name")

	schema, ok := Cruds[table]
	if !ok {
		xhttp.ResponseJson(w, map[string]any{
			"code":    400,
			"message": "not support collection",
		})
		return
	}

	requestBody, err := xhttp.DecodeBody[xdb.Record](r)
	if err != nil {
		xhttp.ResponseJson(w, map[string]any{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	m := xdb.New(table)
	if schema.NewModel != nil {
		m = schema.NewModel(r)
	}

	if schema.BeforeCreate != nil {
		_requestBody, err := schema.BeforeCreate(r, *requestBody)
		if err != nil {
			xhttp.ResponseJson(w, map[string]any{
				"code":    500,
				"message": err.Error(),
			})
			return
		}
		requestBody = &_requestBody
	}

	id, err := m.Ctx(r.Context()).Insert(*requestBody)
	if err != nil {
		xhttp.ResponseJson(w, map[string]any{
			"code":    500,
			"message": err.Error(),
		})
		return
	}
	xhttp.ResponseJson(w, map[string]any{
		"code": 0,
		"msg":  "success",
		"data": map[string]any{
			"id": id,
		},
	})
}

func Get(w http.ResponseWriter, r *http.Request) {
	table := xhttp.Vars(r, "table_name")
	id := xhttp.Vars(r, "id")

	schema, ok := Cruds[table]
	if !ok {
		xhttp.ResponseJson(w, map[string]any{
			"code":    400,
			"message": "not support collection",
		})
		return
	}

	opt := []xdb.Option{
		xdb.WhereEq("id", id),
	}
	if schema.BeforeGet != nil {
		opt = schema.BeforeGet(r, opt)
	}

	m := xdb.New(table)
	if schema.NewModel != nil {
		m = schema.NewModel(r)
	}

	row := m.Ctx(r.Context()).SelectOne(opt...)
	if row.Err != nil {
		xhttp.ResponseJson(w, map[string]any{
			"code":    500,
			"message": row.Err.Error(),
		})
		return
	}

	xhttp.ResponseJson(w, map[string]any{
		"code":    0,
		"message": "success",
		"data":    row.Data,
	})
}

func Update(w http.ResponseWriter, r *http.Request) {
	table := xhttp.Vars(r, "table_name")
	id := xhttp.Vars(r, "id")

	schema, ok := Cruds[table]
	if !ok {
		xhttp.ResponseJson(w, map[string]any{
			"code":    400,
			"message": "not support collection",
		})
		return
	}

	requestBody, err := xhttp.DecodeBody[xdb.Record](r)
	if err != nil {
		xhttp.ResponseJson(w, map[string]any{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	updateData := xdb.Record{}
	_schema := schema.DecodeSchema()
	for _, field := range _schema.FormItems {
		if val, ok := (*requestBody)[field.Field]; ok {
			if field.Field == "id" {
				continue
			}
			updateData[field.Field] = val
		}
	}

	opt := []xdb.Option{
		xdb.WhereEq("id", id),
	}
	if schema.BeforeUpdate != nil {
		_updateData, err := schema.BeforeUpdate(r, updateData)
		if err != nil {
			xhttp.ResponseJson(w, map[string]any{
				"code":    200,
				"message": err.Error(),
			})
			return
		}
		updateData = _updateData
	}

	m := xdb.New(table)
	if schema.NewModel != nil {
		m = schema.NewModel(r)
	}

	_, err = m.Ctx(r.Context()).Update(updateData, opt...)
	if err != nil {
		xhttp.ResponseJson(w, map[string]any{
			"code":    500,
			"message": err.Error(),
		})
		return
	}
	xhttp.ResponseJson(w, map[string]any{
		"code": 0,
		"msg":  "success",
	})
}

func Delete(w http.ResponseWriter, r *http.Request) {
	table := xhttp.Vars(r, "table_name")
	id := xhttp.Vars(r, "id")

	schema, ok := Cruds[table]
	if !ok {
		xhttp.ResponseJson(w, map[string]any{
			"code":    400,
			"message": "not support collection",
		})
		return
	}

	opt := []xdb.Option{
		xdb.WhereEq("id", id),
	}
	if schema.BeforeDelete != nil {
		opt = schema.BeforeDelete(r, opt)
	}

	m := xdb.New(table)
	if schema.NewModel != nil {
		m = schema.NewModel(r)
	}

	_, err := m.Ctx(r.Context()).Update(xdb.Record{"is_deleted": 1}, opt...)
	if err != nil {
		xhttp.ResponseJson(w, map[string]any{
			"code":    500,
			"message": err.Error(),
		})
		return
	}
	xhttp.ResponseJson(w, map[string]any{
		"code": 0,
		"msg":  "success",
	})
}

func Options(w http.ResponseWriter, r *http.Request) {
	table := xhttp.Vars(r, "table_name")

	rule, ok := Cruds[table]
	if !ok {
		xhttp.ResponseJson(w, map[string]any{
			"code":    400,
			"message": "not support collection",
		})
		return
	}

	kw := r.URL.Query().Get("kw")
	field := r.URL.Query().Get("field")
	operator := r.URL.Query().Get("operator")
	if operator == "" {
		operator = "like" // default operator
	}

	// Initialize model based on rules
	model := xdb.New(table)
	if rule.NewModel != nil {
		model = rule.NewModel(r)
	}

	// Building query options
	var options []xdb.Option
	if match, _ := regexp.MatchString(`^\d+$`, kw); match {
		options = append(options, xdb.WhereEq("id", kw))
	} else if field != "" {
		switch operator {
		case "like":
			options = append(options, xdb.WhereLike(field, "%"+kw+"%"))
		case "eq":
			options = append(options, xdb.WhereEq(field, kw))
		}
	}

	// Perform the query for multiple records
	rows, err := model.Selects(options...)
	if err != nil {
		xhttp.ResponseJson(w, map[string]any{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	xhttp.ResponseJson(w, map[string]any{
		"code": 0,
		"msg":  "success",
		"data": rows,
	})
}
