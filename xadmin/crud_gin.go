package xadmin

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"

	"github.com/daodao97/xgo/xdb"
)

func GinPageSchema(c *gin.Context) {
	table := c.Param("table_name")

	schema, ok := Cruds[table]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "not support collection",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": schema.GetSchema(),
	})
}

func GinList(c *gin.Context) {
	table := c.Param("table_name")

	schema, ok := Cruds[table]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "not support collection",
		})
		return
	}

	ps := c.DefaultQuery("_ps", "20")
	pn := c.DefaultQuery("_pn", "1")

	m := xdb.New(table)
	if schema.NewModel != nil {
		m = schema.NewModel(c.Request)
	}

	_schema := schema.DecodeSchema()
	var opt []xdb.Option
	for _, field := range _schema.Filter {
		value := c.Query(field.Field)
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
		opt = schema.BeforeList(c.Request, opt)
	}

	count, err := m.Ctx(c).Count(opt...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	if count == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "success",
			"data": gin.H{
				"list": []any{},
				"page": gin.H{
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

	sortBy := c.Query("_sort_by")
	sortType := c.DefaultQuery("_sort_type", "desc")
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

	records, err := m.Ctx(c).Selects(opt...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	list := make([]xdb.Record, 0)
	if len(records) > 0 {
		list = records
	}

	if schema.AfterList != nil {
		list = schema.AfterList(c.Request, list)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{
			"list": list,
			"page": gin.H{
				"pn":    cast.ToInt(pn),
				"ps":    cast.ToInt(ps),
				"total": count,
			},
		},
	})
}

func GinCreate(c *gin.Context) {
	table := c.Param("table_name")

	schema, ok := Cruds[table]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "not support collection",
		})
		return
	}

	var requestBody xdb.Record
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	m := xdb.New(table)
	if schema.NewModel != nil {
		m = schema.NewModel(c.Request)
	}

	id, err := m.Ctx(c).Insert(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{
			"id": id,
		},
	})
}

func GinGet(c *gin.Context) {
	table := c.Param("table_name")
	id := c.Param("id")

	schema, ok := Cruds[table]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "not support collection",
		})
		return
	}

	opt := []xdb.Option{
		xdb.WhereEq("id", id),
	}
	if schema.BeforeGet != nil {
		opt = schema.BeforeGet(c.Request, opt)
	}

	m := xdb.New(table)
	if schema.NewModel != nil {
		m = schema.NewModel(c.Request)
	}

	row := m.Ctx(c).SelectOne(opt...)
	if row.Err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": row.Err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    row.Data,
	})
}

func GinUpdate(c *gin.Context) {
	table := c.Param("table_name")
	id := c.Param("id")

	schema, ok := Cruds[table]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "not support collection",
		})
		return
	}

	var requestBody xdb.Record
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	updateData := xdb.Record{}
	_schema := schema.DecodeSchema()
	for _, field := range _schema.FormItems {
		if val, ok := requestBody[field.Field]; ok {
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
		updateData = schema.BeforeUpdate(c.Request, updateData)
	}

	m := xdb.New(table)
	if schema.NewModel != nil {
		m = schema.NewModel(c.Request)
	}

	_, err := m.Ctx(c).Update(updateData, opt...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
	})
}

func GinDelete(c *gin.Context) {
	table := c.Param("table_name")
	id := c.Param("id")

	schema, ok := Cruds[table]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "not support collection",
		})
		return
	}

	opt := []xdb.Option{
		xdb.WhereEq("id", id),
	}
	if schema.BeforeDelete != nil {
		opt = schema.BeforeDelete(c.Request, opt)
	}

	m := xdb.New(table)
	if schema.NewModel != nil {
		m = schema.NewModel(c.Request)
	}

	_, err := m.Ctx(c).Update(xdb.Record{"is_deleted": 1}, opt...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
	})
}

func GinOptions(c *gin.Context) {
	table := c.Param("table_name")

	rule, ok := Cruds[table]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "not support collection",
		})
		return
	}

	kw := c.Query("kw")
	field := c.Query("field")
	operator := c.DefaultQuery("operator", "like")

	// Initialize model based on rules
	model := xdb.New(table)
	if rule.NewModel != nil {
		model = rule.NewModel(c.Request)
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": rows,
	})
}
