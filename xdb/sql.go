package xdb

import (
	"fmt"
	"strings"
)

const selectMod = "select %s from %s"
const insertMod = "insert into %s (%s) values (%s)"
const updateMod = "update %s set %s"
const deleteMod = "delete from %s"

type Option = func(opts *Options)

type Options struct {
	database  string
	table     string
	field     []string
	where     []where
	orderBy   []string
	groupBy   string
	limit     int
	offset    int
	value     []any
	forUpdate bool
}

func table(table string) Option {
	return func(opts *Options) {
		opts.table = table
	}
}

func database(database string) Option {
	return func(opts *Options) {
		opts.database = database
	}
}

func ForUpdate() Option {
	return func(opts *Options) {
		opts.forUpdate = true
	}
}

func Offset(offset int) Option {
	return func(opts *Options) {
		opts.offset = offset
	}
}

func Limit(offset int) Option {
	return func(opts *Options) {
		opts.limit = offset
	}
}

func Pagination(pageNumber, pageSize int) []Option {
	return []Option{
		Limit(pageSize),
		Offset((pageNumber - 1) * pageSize),
	}
}

func Field(name ...string) Option {
	var _name []string
	for _, v := range name {
		if strings.Contains(v, " as ") {
			tmp := strings.Split(v, " as ")
			_name = append(_name, strings.Trim(tmp[0], " ")+" as "+strings.Trim(tmp[1], " "))
		} else if strings.Contains(v, " AS ") {
			tmp := strings.Split(v, " AS ")
			_name = append(_name, strings.Trim(tmp[0], " ")+" as "+strings.Trim(tmp[1], " "))
		} else {
			_name = append(_name, v)
		}
	}
	return func(opts *Options) {
		opts.field = _name
	}
}

func FieldRaw(name string) Option {
	return func(opts *Options) {
		opts.field = append(opts.field, name)
	}
}

func AggregateSum(name string) Option {
	return Field("sum(" + name + ") as aggregate")
}

func AggregateCount(name string) Option {
	return FieldRaw("count(" + name + ") as count")
}

func AggregateMax(name string) Option {
	return Field("max(" + name + ") as aggregate")
}

func Value(val ...any) Option {
	return func(opts *Options) {
		opts.value = val
	}
}

type where struct {
	field    string
	operator string
	value    any
	logic    string
	sub      []where
	raw      string
}

func WhereRaw(raw string) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			raw: raw,
		})
	}
}

func Where(field, operator string, value any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: operator,
			value:    value,
			logic:    "and",
		})
	}
}

func WhereEq(field string, value any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: "=",
			value:    value,
		})
	}
}

func WhereNotEq(field string, value any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: "!=",
			value:    value,
		})
	}
}

func WhereGt(field string, value any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: ">",
			value:    value,
		})
	}
}

// Deprecated: Use WhereGte instead.
func WhereGe(field string, value any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: ">=",
			value:    value,
		})
	}
}

func WhereGte(field string, value any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: ">=",
			value:    value,
		})
	}
}

func WhereLt(field string, value any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: "<",
			value:    value,
		})
	}
}

// Deprecated: Use WhereLte instead.
func WhereLe(field string, value any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: "<=",
			value:    value,
		})
	}
}

func WhereLte(field string, value any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: "<=",
			value:    value,
		})
	}
}

func WhereIn(field string, value []any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: "in",
			value:    value,
		})
	}
}

func WhereNotIn(field string, value []any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: "not in",
			value:    value,
		})
	}
}

func WhereOr(field, operator string, value any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: operator,
			value:    value,
			logic:    "or",
		})
	}
}

func WhereOrEq(field string, value any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: "=",
			value:    value,
			logic:    "or",
		})
	}
}

func WhereOrNotEq(field string, value any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: "!=",
			value:    value,
			logic:    "or",
		})
	}
}

func WhereOrGt(field string, value any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: ">",
			value:    value,
			logic:    "or",
		})
	}
}

func WhereOrGe(field string, value any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: ">=",
			value:    value,
			logic:    "or",
		})
	}
}

func WhereOrLt(field string, value any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: "<",
			value:    value,
			logic:    "or",
		})
	}
}

func WhereOrLe(field string, value any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: "<=",
			value:    value,
			logic:    "or",
		})
	}
}

func WhereOrIn(field string, value []any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: "in",
			value:    value,
			logic:    "or",
		})
	}
}

func WhereOrNotIn(field string, value []any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: "not in",
			value:    value,
			logic:    "or",
		})
	}
}

func WhereGroup(opts ...Option) Option {
	opt := &Options{}
	for _, v := range opts {
		v(opt)
	}
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			logic: "and",
			sub:   opt.where,
		})
	}
}

func WhereOrGroup(opts ...Option) Option {
	opt := &Options{}
	for _, v := range opts {
		v(opt)
	}
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			logic: "or",
			sub:   opt.where,
		})
	}
}

func WhereLike(field string, value any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: "like",
			value:    value,
		})
	}
}

func WhereOrLike(field string, value any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: "like",
			value:    value,
			logic:    "or",
		})
	}
}

func WhereOrNotLike(field string, value any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: "not like",
			value:    value,
		})
	}
}

func WhereBetween(field string, value1 any, value2 any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: "between",
			value:    []any{value1, value2},
		})
	}
}

func WhereFindInSet(field string, value any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: "find_in_set",
			value:    value,
		})
	}
}

func WhereOrFindInSet(field string, value any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: "find_in_set",
			value:    value,
			logic:    "or",
		})
	}
}

func OrderByDesc(field string) Option {
	return func(opts *Options) {
		opts.orderBy = append(opts.orderBy, field+" desc")
	}
}

func OrderByAsc(field string) Option {
	return func(opts *Options) {
		opts.orderBy = append(opts.orderBy, field+" asc")
	}
}

func GroupBy(field string) Option {
	return func(opts *Options) {
		opts.groupBy = field
	}
}

func whereBuilder(condition []where) (sql string, args []any) {
	if len(condition) == 0 {
		return "", nil
	}
	var tokens []string
	for i, v := range condition {
		if i != 0 {
			if v.logic != "" {
				tokens = append(tokens, v.logic)
			} else {
				tokens = append(tokens, "and")
			}
		}

		if v.raw != "" {
			tokens = append(tokens, v.raw)
			continue
		}

		if v.field != "" {
			switch v.operator {
			case "in", "not in":
				val := v.value.([]any)
				var placeholder []string
				for range val {
					placeholder = append(placeholder, "?")
				}
				tokens = append(tokens, fmt.Sprintf("%s %s (%s)", v.field, v.operator, strings.Join(placeholder, ",")))
				args = append(args, val...)
			case "between":
				val := v.value.([]any)
				tokens = append(tokens, fmt.Sprintf("%s %s ? and ?", v.field, v.operator))
				args = append(args, val...)
			case "find_in_set":
				tokens = append(tokens, fmt.Sprintf("find_in_set(?, %s)", v.field))
				args = append(args, v.value)
			default:
				tokens = append(tokens, fmt.Sprintf("%s %s ?", v.field, v.operator))
				args = append(args, v.value)
			}
		}

		if v.sub != nil {
			_sql, _args := whereBuilder(v.sub)
			tokens = append(tokens, "("+_sql+")")
			args = append(args, _args...)
		}
	}

	return strings.Join(tokens, " "), args
}

func SelectBuilder(opts ...Option) (sql string, args []any) {
	_opts := &Options{}
	for _, v := range opts {
		v(_opts)
	}

	_where, args := whereBuilder(_opts.where)
	_field := "*"

	if len(_opts.field) > 0 {
		_field = strings.Join(_opts.field, ", ")
	}

	sql = fmt.Sprintf(selectMod, _field, _opts.table)

	if _where != "" {
		sql = sql + " where " + _where
	}

	if _opts.groupBy != "" {
		sql = sql + " group by " + _opts.groupBy
	}

	if len(_opts.orderBy) > 0 {
		sql = sql + " order by " + strings.Join(_opts.orderBy, ", ")
	}

	if _opts.limit != 0 {
		sql = sql + " limit ? offset ?"
		args = append(args, _opts.limit, _opts.offset)
	}

	if _opts.forUpdate {
		sql = sql + " for update"
	}

	return sql, args
}

func getTable(opt *Options) string {
	if opt.database == "" {
		return opt.table
	}
	return opt.database + "." + opt.table
}

func InsertBuilder(opts ...Option) (sql string, args []any) {
	_opts := &Options{}
	for _, v := range opts {
		v(_opts)
	}
	var _val []string
	for range _opts.field {
		_val = append(_val, "?")
	}
	sql = fmt.Sprintf(insertMod, getTable(_opts), strings.Join(_opts.field, ", "), strings.Join(_val, ","))
	args = _opts.value
	return sql, args
}

func HaveFieldInWhere(field string, opts ...Option) (any, bool) {
	_opts := &Options{}
	for _, v := range opts {
		v(_opts)
	}
	for _, v := range _opts.where {
		if field == v.field {
			return v.value, true
		}
	}
	return nil, false
}

func FieldsInWhere(opts ...Option) []string {
	_opts := &Options{}
	for _, v := range opts {
		v(_opts)
	}
	var fields []string
	for _, v := range _opts.where {
		fields = append(fields, v.field)
	}
	return fields
}

func UniqueString(str []string) []string {
	// 使用 map 来去重
	uniqueMap := make(map[string]struct{})
	var result []string

	// 遍历输入的字符串切片，将唯一的值添加到结果中
	for _, s := range str {
		if _, exists := uniqueMap[s]; !exists {
			uniqueMap[s] = struct{}{}
			result = append(result, s)
		}
	}

	return result
}

func UpdateBuilder(opts ...Option) (sql string, args []any) {
	_opts := &Options{}
	for _, v := range opts {
		v(_opts)
	}
	var _set []string
	for i, v := range _opts.field {
		_set = append(_set, parseSet(v, _opts.value[i]))
	}
	sql = fmt.Sprintf(updateMod, getTable(_opts), strings.Join(_set, ","))
	args = parseSetValues(_opts.value)
	if len(_opts.where) > 0 {
		_where, _args := whereBuilder(_opts.where)
		sql = sql + " where " + _where
		args = append(args, _args...)
	}
	return sql, args
}

func DeleteBuilder(opts ...Option) (sql string, args []any) {
	_opts := &Options{}
	for _, v := range opts {
		v(_opts)
	}
	sql = fmt.Sprintf(deleteMod, _opts.table)
	if len(_opts.where) > 0 {
		_where, _args := whereBuilder(_opts.where)
		sql = sql + " where " + _where
		args = append(args, _args...)
	}
	return sql, args
}

// insert or update value
const (
	OpAdd = "+"
	OpSub = "-"
)

type UpdateValue struct {
	Value interface{}
	Op    string // 操作符：+, -
}

func SelfAdd(value any) UpdateValue {
	return UpdateValue{Value: value, Op: OpAdd}
}

func SelfSub(value any) UpdateValue {
	return UpdateValue{Value: value, Op: OpSub}
}

func parseSet(field string, value any) string {
	if uv, ok := value.(UpdateValue); ok {
		switch uv.Op {
		case OpAdd:
			return fmt.Sprintf("%s = %s + ?", field, field)
		case OpSub:
			return fmt.Sprintf("%s = %s - ?", field, field)
		}
	}
	return fmt.Sprintf("%s = ?", field)
}

func parseSetValues(values []any) []any {
	for i, v := range values {
		if uv, ok := v.(UpdateValue); ok {
			values[i] = uv.Value
		}
	}
	return values
}

func parseSetValue(value any) any {
	if uv, ok := value.(UpdateValue); ok {
		return uv.Value
	}
	return value
}
