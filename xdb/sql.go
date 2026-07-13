package xdb

import (
	"fmt"
	"strings"
	"unicode"
)

const selectMod = "select %s from %s"
const insertMod = "insert into %s (%s) values (%s)"
const updateMod = "update %s set %s"
const deleteMod = "delete from %s"

type Option = func(opts *Options)

type sqlFragment struct {
	sql  string
	args []any
	raw  bool
}

type Options struct {
	database  string
	table     string
	field     []string
	rawField  map[string]bool
	where     []where
	orderBy   []sqlFragment
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
		if opts.rawField == nil {
			opts.rawField = make(map[string]bool)
		}
		opts.rawField[name] = true
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
	rawArgs  []any
}

func WhereRaw(raw string) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			raw: raw,
		})
	}
}

func WhereOrRaw(raw string) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			raw:   raw,
			logic: "or",
		})
	}
}

// WhereRawArgs 追加一个带 ? 占位符的 raw WHERE 片段。
// raw 中每个 ? 按从左到右顺序消费 args 中的一个值，作为绑定参数下发
// （driver 预处理占位符，非字符串拼接）。
// 约定 raw 中 ? 的个数应等于 len(args)，由调用方保证；数量不匹配时由 driver 在执行阶段报错。
// WhereRaw(raw) 等价于 WhereRawArgs(raw) 无参数。
func WhereRawArgs(raw string, args ...any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			raw:     raw,
			rawArgs: args,
		})
	}
}

// WhereOrRawArgs 是 WhereRawArgs 的 OR 逻辑变体。
func WhereOrRawArgs(raw string, args ...any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			raw:     raw,
			rawArgs: args,
			logic:   "or",
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

func WhereIsNil(field string) Option {
	return WhereRaw(field + " is null")
}

func WhereNotNil(field string) Option {
	return WhereRaw(field + " is not null")
}

func WhereOrIsNil(field string) Option {
	return WhereOrRaw(field + " is null")
}

func WhereOrNotNil(field string) Option {
	return WhereOrRaw(field + " is not null")
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

// WhereFindInSet adds a FIND_IN_SET condition to the query.
// Note: This function is MySQL-specific and will not work with PostgreSQL or SQLite.
// For cross-database compatibility, consider using alternative approaches like JSON columns or separate junction tables.
func WhereFindInSet(field string, value any) Option {
	return func(opts *Options) {
		opts.where = append(opts.where, where{
			field:    field,
			operator: "find_in_set",
			value:    value,
		})
	}
}

// WhereOrFindInSet adds an OR FIND_IN_SET condition to the query.
// Note: This function is MySQL-specific and will not work with PostgreSQL or SQLite.
// For cross-database compatibility, consider using alternative approaches like JSON columns or separate junction tables.
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
		opts.orderBy = append(opts.orderBy, sqlFragment{sql: field + " desc"})
	}
}

func OrderByAsc(field string) Option {
	return func(opts *Options) {
		opts.orderBy = append(opts.orderBy, sqlFragment{sql: field + " asc"})
	}
}

// OrderByRaw appends a raw ORDER BY expression with bound arguments.
// raw must be trusted application SQL; dynamic values belong in args.
// Placeholder mismatches are reported by the database driver at execution time.
func OrderByRaw(raw string, args ...any) Option {
	return func(opts *Options) {
		opts.orderBy = append(opts.orderBy, sqlFragment{
			sql:  raw,
			args: args,
			raw:  true,
		})
	}
}

func orderByBuilder(dialect Dialect, orderBy []sqlFragment) (sql string, args []any) {
	expressions := make([]string, 0, len(orderBy))
	for _, fragment := range orderBy {
		expression := fragment.sql
		if !fragment.raw {
			expression = quoteOrderItem(dialect, expression)
		}
		expressions = append(expressions, expression)
		args = append(args, fragment.args...)
	}
	return strings.Join(expressions, ", "), args
}

func GroupBy(field string) Option {
	return func(opts *Options) {
		opts.groupBy = field
	}
}

func quoteIdentifier(dialect Dialect, identifier string) string {
	if dialect == nil {
		return identifier
	}
	if !isSimpleIdentifierPath(identifier) {
		return identifier
	}
	if quoter, ok := dialect.(identifierQuoter); ok {
		return quoter.QuoteIdentifier(identifier)
	}
	return identifier
}

func quoteIdentifiers(dialect Dialect, identifiers []string) []string {
	quoted := make([]string, 0, len(identifiers))
	for _, identifier := range identifiers {
		quoted = append(quoted, quoteIdentifier(dialect, identifier))
	}
	return quoted
}

func quoteSelectFields(dialect Dialect, fields []string) string {
	return quoteSelectFieldsWithRaw(dialect, fields, nil)
}

func quoteSelectFieldsWithRaw(dialect Dialect, fields []string, raw map[string]bool) string {
	quoted := make([]string, 0, len(fields))
	for _, field := range fields {
		if raw[field] {
			quoted = append(quoted, field)
			continue
		}
		quoted = append(quoted, quoteAliasedIdentifier(dialect, field))
	}
	return strings.Join(quoted, ", ")
}

func quoteAliasedIdentifier(dialect Dialect, expr string) string {
	left, alias, ok := splitAlias(expr)
	if !ok {
		return quoteIdentifier(dialect, strings.TrimSpace(expr))
	}
	left = strings.TrimSpace(left)
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return quoteIdentifier(dialect, left)
	}
	return quoteIdentifier(dialect, left) + " as " + quoteIdentifier(dialect, alias)
}

func splitAlias(expr string) (left, alias string, ok bool) {
	lower := strings.ToLower(expr)
	index := strings.LastIndex(lower, " as ")
	if index < 0 {
		return "", "", false
	}
	return expr[:index], expr[index+4:], true
}

func quoteIdentifierList(dialect Dialect, identifiers string) string {
	parts := strings.Split(identifiers, ",")
	for i, part := range parts {
		parts[i] = quoteAliasedIdentifier(dialect, strings.TrimSpace(part))
	}
	return strings.Join(parts, ", ")
}

func quoteOrderItem(dialect Dialect, order string) string {
	parts := strings.Fields(order)
	if len(parts) == 0 {
		return order
	}
	if len(parts) == 1 {
		return quoteIdentifier(dialect, parts[0])
	}
	direction := strings.ToLower(parts[len(parts)-1])
	if direction != "asc" && direction != "desc" {
		return order
	}
	identifier := strings.Join(parts[:len(parts)-1], " ")
	return quoteIdentifier(dialect, identifier) + " " + direction
}

func isSimpleIdentifierPath(identifier string) bool {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" || identifier == "*" {
		return true
	}
	if isQuotedIdentifier(identifier) {
		return true
	}
	parts := strings.Split(identifier, ".")
	for _, part := range parts {
		if part == "" {
			return false
		}
		if part == "*" || isQuotedIdentifier(part) {
			continue
		}
		for _, r := range part {
			if r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
				continue
			}
			return false
		}
	}
	return true
}

func validateIdentifier(identifier string) error {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return nil
	}
	if !isSimpleIdentifierPath(identifier) {
		return fmt.Errorf("%w: %s", ErrInvalidIdentifier, identifier)
	}
	return nil
}

func validateAliasedIdentifier(identifier string) error {
	left, alias, ok := splitAlias(identifier)
	if !ok {
		return validateIdentifier(identifier)
	}
	if err := validateIdentifier(left); err != nil {
		return err
	}
	return validateIdentifier(alias)
}

func validateOrderIdentifier(identifier string) error {
	parts := strings.Fields(identifier)
	if len(parts) == 0 {
		return nil
	}
	if len(parts) == 1 {
		return validateIdentifier(parts[0])
	}
	direction := strings.ToLower(parts[len(parts)-1])
	if direction != "asc" && direction != "desc" {
		return fmt.Errorf("%w: %s", ErrInvalidIdentifier, identifier)
	}
	return validateIdentifier(strings.Join(parts[:len(parts)-1], " "))
}

func validateWhereIdentifiers(condition []where) error {
	for _, item := range condition {
		if item.raw != "" {
			continue
		}
		if item.field != "" {
			if err := validateIdentifier(item.field); err != nil {
				return err
			}
		}
		if len(item.sub) > 0 {
			if err := validateWhereIdentifiers(item.sub); err != nil {
				return err
			}
		}
	}
	return nil
}

func whereBuilder(condition []where) (sql string, args []any) {
	return whereBuilderWithDialect(nil, condition)
}

func whereBuilderWithDialect(dialect Dialect, condition []where) (sql string, args []any) {
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
			args = append(args, v.rawArgs...)
			continue
		}

		if v.field != "" {
			switch v.operator {
			case "in", "not in":
				val := v.value.([]any)
				if len(val) == 0 {
					if v.operator == "in" {
						tokens = append(tokens, "1 = 0")
					} else {
						tokens = append(tokens, "1 = 1")
					}
					continue
				}
				var placeholder []string
				for range val {
					placeholder = append(placeholder, "?")
				}
				tokens = append(tokens, fmt.Sprintf("%s %s (%s)", quoteIdentifier(dialect, v.field), v.operator, strings.Join(placeholder, ",")))
				args = append(args, val...)
			case "between":
				val := v.value.([]any)
				tokens = append(tokens, fmt.Sprintf("%s %s ? and ?", quoteIdentifier(dialect, v.field), v.operator))
				args = append(args, val...)
			case "find_in_set":
				tokens = append(tokens, fmt.Sprintf("find_in_set(?, %s)", quoteIdentifier(dialect, v.field)))
				args = append(args, v.value)
			default:
				tokens = append(tokens, fmt.Sprintf("%s %s ?", quoteIdentifier(dialect, v.field), v.operator))
				args = append(args, v.value)
			}
		}

		if v.sub != nil {
			_sql, _args := whereBuilderWithDialect(dialect, v.sub)
			tokens = append(tokens, "("+_sql+")")
			args = append(args, _args...)
		}
	}

	return strings.Join(tokens, " "), args
}

func SelectBuilder(opts ...Option) (sql string, args []any) {
	return SelectBuilderWithDialect(nil, opts...)
}

func SelectBuilderWithDialect(dialect Dialect, opts ...Option) (sql string, args []any) {
	_opts := &Options{}
	for _, v := range opts {
		v(_opts)
	}

	_where, args := whereBuilderWithDialect(dialect, _opts.where)
	_field := "*"

	if len(_opts.field) > 0 {
		_field = quoteSelectFieldsWithRaw(dialect, _opts.field, _opts.rawField)
	}

	sql = fmt.Sprintf(selectMod, _field, quoteTable(dialect, _opts))

	if _where != "" {
		sql = sql + " where " + _where
	}

	if _opts.groupBy != "" {
		sql = sql + " group by " + quoteIdentifierList(dialect, _opts.groupBy)
	}

	if len(_opts.orderBy) > 0 {
		_orderBy, _args := orderByBuilder(dialect, _opts.orderBy)
		sql = sql + " order by " + _orderBy
		args = append(args, _args...)
	}

	if _opts.limit != 0 {
		if dialect != nil {
			sql = sql + dialect.LimitOffset(_opts.limit, _opts.offset)
		} else {
			sql = sql + " limit ? offset ?"
		}
		args = append(args, _opts.limit, _opts.offset)
	}

	if _opts.forUpdate {
		sql = sql + forUpdateSQL(dialect)
	}

	return sql, args
}

func getTable(opt *Options) string {
	return quoteTable(nil, opt)
}

func quoteTable(dialect Dialect, opt *Options) string {
	if opt.database == "" {
		return quoteIdentifier(dialect, opt.table)
	}
	return quoteIdentifier(dialect, opt.database) + "." + quoteIdentifier(dialect, opt.table)
}

func InsertBuilder(opts ...Option) (sql string, args []any) {
	return InsertBuilderWithDialect(nil, opts...)
}

func InsertBuilderWithDialect(dialect Dialect, opts ...Option) (sql string, args []any) {
	_opts := &Options{}
	for _, v := range opts {
		v(_opts)
	}
	var _val []string
	for range _opts.field {
		_val = append(_val, "?")
	}
	sql = fmt.Sprintf(insertMod, quoteTable(dialect, _opts), quoteIdentifierList(dialect, strings.Join(_opts.field, ", ")), strings.Join(_val, ","))
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
	return UpdateBuilderWithDialect(nil, opts...)
}

func UpdateBuilderWithDialect(dialect Dialect, opts ...Option) (sql string, args []any) {
	_opts := &Options{}
	for _, v := range opts {
		v(_opts)
	}
	var _set []string
	for i, v := range _opts.field {
		_set = append(_set, parseSetWithDialect(dialect, v, _opts.value[i]))
	}
	sql = fmt.Sprintf(updateMod, quoteTable(dialect, _opts), strings.Join(_set, ","))
	args = parseSetValues(_opts.value)
	if len(_opts.where) > 0 {
		_where, _args := whereBuilderWithDialect(dialect, _opts.where)
		sql = sql + " where " + _where
		args = append(args, _args...)
	}
	return sql, args
}

func DeleteBuilder(opts ...Option) (sql string, args []any) {
	return DeleteBuilderWithDialect(nil, opts...)
}

func DeleteBuilderWithDialect(dialect Dialect, opts ...Option) (sql string, args []any) {
	_opts := &Options{}
	for _, v := range opts {
		v(_opts)
	}
	sql = fmt.Sprintf(deleteMod, quoteTable(dialect, _opts))
	if len(_opts.where) > 0 {
		_where, _args := whereBuilderWithDialect(dialect, _opts.where)
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
	return parseSetWithDialect(nil, field, value)
}

func parseSetWithDialect(dialect Dialect, field string, value any) string {
	quotedField := quoteIdentifier(dialect, field)
	if uv, ok := value.(UpdateValue); ok {
		switch uv.Op {
		case OpAdd:
			return fmt.Sprintf("%s = %s + ?", quotedField, quotedField)
		case OpSub:
			return fmt.Sprintf("%s = %s - ?", quotedField, quotedField)
		}
	}
	return fmt.Sprintf("%s = ?", quotedField)
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
