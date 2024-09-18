package xdb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/spf13/cast"

	"github.com/daodao97/xgo/xlog"

	"github.com/daodao97/xgo/xdb/interval/util"
)

var ErrNotFound = errors.New("record not found")

type Record map[string]interface{}

func (r Record) Binding(dest interface{}) error {
	if !util.AllowType(dest, []string{"*struct", "**struct"}) {
		return ErrRowBindingType
	}

	return util.Binding(r, dest)
}

func (r Record) Get(key string) (interface{}, bool) {
	v, ok := r[key]
	return v, ok
}

func (r Record) GetString(key string) string {
	v, ok := r[key]
	if !ok {
		return ""
	}
	return cast.ToString(v)
}

func (r Record) GetInt(key string) int {
	v, ok := r[key]
	if !ok {
		return 0
	}
	return cast.ToInt(v)
}

func (r Record) GetArray(key string) []any {
	v, ok := r[key]
	if !ok {
		return []any{}
	}
	return cast.ToSlice(v)
}

func (r Record) GetTime(key string) *time.Time {
	v, ok := r[key]
	if !ok {
		return nil
	}
	return v.(*time.Time)
}

func (r Record) GetTimeFormat(key string, format string) string {
	v, ok := r[key]
	if !ok {
		return ""
	}

	t, ok := v.(time.Time)
	if !ok {
		return ""
	}
	return t.Format(format)
}

func (r Record) GetAny(key string) any {
	v, ok := r[key]
	if !ok {
		return nil
	}
	return v
}

func (r Record) GetBool(key string) bool {
	v, ok := r[key]
	if !ok {
		return false
	}
	return cast.ToBool(v)
}

func (r Record) GetFloat64(key string) float64 {
	v, ok := r[key]
	if !ok {
		return 0
	}
	return cast.ToFloat64(v)
}

func (r Record) GetRecord(key string) Record {
	v, ok := r[key]
	if !ok {
		return nil
	}

	var record Record

	mapstructure.Decode(v, &record)

	return record
}

type Model interface {
	PrimaryKey() string
	//Deprecated: use Selects instead
	Select(opt ...Option) (rows *Rows)
	//Deprecated: use Single instead
	SelectOne(opt ...Option) *Row
	Selects(opt ...Option) ([]Record, error)
	Single(opt ...Option) (Record, error)
	Count(opt ...Option) (count int64, err error)
	Insert(record Record) (lastId int64, err error)
	Update(record Record, opt ...Option) (ok bool, err error)
	InsertOrUpdate(record Record, updateFields ...string) (resp Record, affected int64, err error)
	Delete(opt ...Option) (ok bool, err error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	//Deprecated: use FindById instead
	FindBy(id string) *Row
	//Deprecated: use FindByField instead
	FindByKey(key string, val string) *Row
	FindById(id string) (Record, error)
	FindByField(field string, val string) (Record, error)
	UpdateBy(id string, record Record) (bool, error)
	Ctx(ctx context.Context) Model
}

type model struct {
	connection      string
	database        string
	table           string
	fakeDelKey      string
	primaryKey      string
	cacheKey        []string
	columnHook      map[string]HookData
	columnValidator []Valid
	hasOne          []HasOpts
	hasMany         []HasOpts
	client          *sql.DB
	readClient      *sql.DB
	config          *Config
	saveZero        bool
	enableValidator bool
	err             error
	ctx             context.Context
}

func New(table string, baseOpt ...With) Model {
	m := &model{
		connection: "default",
		primaryKey: "id",
		table:      table,
	}

	if table == "" {
		m.err = errors.New("table name is empty")
		return m
	}

	for _, v := range baseOpt {
		v(m)
	}

	if m.client == nil {
		p, err := db(m.connection)
		if err != nil {
			m.err = err
			return m
		}
		m.client = p.db
		m.config = p.conf
	}
	if m.readClient == nil {
		p, err := db(readConn(m.connection))
		if err == nil {
			m.readClient = p.db
		}
	}
	m.enableValidator = true
	return m
}

func (m *model) Ctx(ctx context.Context) Model {
	m.ctx = ctx
	return m
}

func (m *model) PrimaryKey() string {
	return m.primaryKey
}

func (m *model) Select(opt ...Option) (rows *Rows) {
	var kv []interface{}
	var err error
	defer dbLog(m.ctx, "Select", time.Now(), &err, &kv)

	if m.err != nil {
		err = m.err
		return &Rows{Err: m.err}
	}
	opts := new(Options)
	opt = append(opt, table(m.table), database(m.database))
	if m.fakeDelKey != "" {
		opt = append(opt, WhereEq(m.fakeDelKey, 0))
	}
	for _, o := range opt {
		o(opts)
	}

	_sql, args := SelectBuilder(opt...)

	client := m.client
	if m.readClient != nil {
		client = m.readClient
	}

	res, err := query(client, _sql, args...)
	kv = append(kv, "sql", _sql, "args", args)
	if err != nil {
		return &Rows{Err: err}
	}

	for _, has := range m.hasOne {
		res, err = m.hasOneData(res, has)
		if err != nil {
			return &Rows{Err: err}
		}
	}

	for _, has := range m.hasMany {
		res, err = m.hasManyData(res, has)
		if err != nil {
			return &Rows{Err: err}
		}
	}

	for k, v := range m.columnHook {
		for i, r := range res {
			for field, val := range r.Data {
				if k == field {
					overVal, err1 := v.Output(res[i].Data, val)
					if err1 != nil {
						err = err1
						return &Rows{Err: err}
					}
					res[i].Data[field] = overVal
				}
			}
		}
	}

	// is is set fake del key, delete this field from the record
	if m.fakeDelKey != "" {
		for _, r := range res {
			delete(r.Data, m.fakeDelKey)
		}
	}

	if res == nil {
		res = []Row{}
	}

	return &Rows{List: res, Err: err}
}

func (m *model) Selects(opt ...Option) ([]Record, error) {
	rows := m.Select(opt...)
	if rows.Err != nil {
		return nil, rows.Err
	}
	if len(rows.List) == 0 {
		return []Record{}, nil
	}

	var records []Record
	for _, row := range rows.List {
		records = append(records, row.Data)
	}

	return records, nil
}

func (m *model) SelectOne(opt ...Option) *Row {
	opt = append(opt, Limit(1))
	rows := m.Select(opt...)
	if rows.Err != nil {
		return &Row{Err: rows.Err}
	}
	if len(rows.List) == 0 {
		return &Row{
			Err: ErrNotFound,
		}
	}
	return &rows.List[0]
}

func (m *model) Single(opt ...Option) (Record, error) {
	rows := m.Select(opt...)
	if rows.Err != nil {
		return nil, rows.Err
	}
	if len(rows.List) == 0 {
		return nil, ErrNotFound
	}
	return rows.List[0].Data, nil
}

func (m *model) Count(opt ...Option) (count int64, err error) {
	opt = append(opt, table(m.table), AggregateCount("*"))
	var result struct {
		Count int64
	}
	err = m.SelectOne(opt...).Binding(&result)
	if err != nil {
		return 0, err
	}

	return result.Count, nil
}

func (m *model) Insert(record Record) (lastId int64, err error) {
	if m.err != nil {
		return 0, m.err
	}

	var kv []interface{}
	defer dbLog(m.ctx, "Insert", time.Now(), &err, &kv)

	_record := record
	if len(_record) == 0 {
		return 0, errors.New("empty record to insert, if your record is struct please set xdb tag")
	}

	_record, err = m.hookInput(_record)
	if err != nil {
		return 0, err
	}

	if m.enableValidator {
		for _, v := range m.columnValidator {
			err = v(NewValidOpt(withRow(_record), WithModel(m)))
			if err != nil {
				return 0, err
			}
		}
	}

	delete(_record, m.primaryKey)
	if len(_record) == 0 {
		return 0, errors.New("empty record to insert")
	}

	ks, vs := m.recordToKV(_record)
	_sql, args := InsertBuilder(table(m.table), Field(ks...), Value(vs...))

	if m.config.Driver == "postgres" {
		_sql = _sql + " RETURNING " + m.primaryKey
	}

	kv = append(kv, "sql", _sql, "args", vs)

	if m.config.Driver == "postgres" {
		err = m.client.QueryRow(_sql, args...).Scan(&lastId)
	} else {
		result, err := exec(m.client, _sql, args...)
		if err != nil {
			return 0, err
		}
		return result.LastInsertId()
	}

	if err != nil {
		return 0, err
	}

	return lastId, nil
}

func (m *model) Update(record Record, opt ...Option) (ok bool, err error) {
	if m.err != nil {
		return false, m.err
	}

	var kv []interface{}
	defer dbLog(m.ctx, "Update", time.Now(), &err, &kv)

	_record := record
	if len(_record) == 0 {
		return false, errors.New("empty record to update, if your record is struct please set xdb tag")
	}

	if id, ok := _record[m.primaryKey]; ok {
		kv = append(kv, m.primaryKey, id)
		opt = append(opt, WhereEq(m.primaryKey, id))
	}

	_record, err = m.hookInput(_record)
	if err != nil {
		return false, err
	}

	delete(_record, m.primaryKey)
	if len(_record) == 0 {
		return false, errors.New("empty record to update")
	}

	if m.enableValidator {
		for _, v := range m.columnValidator {
			err = v(NewValidOpt(withRow(_record), WithModel(m)))
			if err != nil {
				return false, err
			}
		}
	}

	ks, vs := m.recordToKV(_record)
	opt = append(opt, table(m.table), Field(ks...), Value(vs...))

	_sql, args := UpdateBuilder(opt...)
	kv = append(kv, "sql", _sql, "args", args)

	result, err := exec(m.client, _sql, args...)
	if err != nil {
		return false, err
	}

	effect, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	cacheKey := append(m.cacheKey, m.primaryKey)
	for _, k := range cacheKey {
		val, ok := HaveFieldInWhere(k, opt...)
		if ok && cache != nil {
			// if update primary key, delete old cache
			if k == m.primaryKey {
				key := m.cacheKeyPrefix(cast.ToString(val))
				err = cache.Del(context.Background(), key)
				if err != nil {
					xlog.Error("del key after update", xlog.Any(k, val), xlog.Err(err))
				} else {
					xlog.Debug("del key after update", xlog.Any(k, val))
				}
			} else {
				// if update other field, delete cache by primary key
				cachedPk, _ := cache.Get(context.Background(), m.cacheKeyPrefix(cast.ToString(val)))
				if cachedPk != "" {
					key := m.cacheKeyPrefix(cachedPk)
					err = cache.Del(context.Background(), key)
					if err != nil {
						xlog.Error("del key after update", xlog.Any(k, val), xlog.Err(err))
					} else {
						xlog.Debug("del key after update", xlog.Any(k, val))
					}
				}
			}
		}
	}

	return effect >= int64(0), nil
}

func (m *model) InsertOrUpdate(record Record, updateFields ...string) (resp Record, affected int64, err error) {
	if m.err != nil {
		return nil, 0, m.err
	}

	var kv []interface{}
	defer dbLog(m.ctx, "InsertOrUpdate", time.Now(), &err, &kv)

	if len(record) == 0 {
		return nil, 0, errors.New("空记录无法插入或更新")
	}

	// 准备插入的字段和值
	var fields []string
	var values []interface{}
	for field, value := range record {
		fields = append(fields, field)
		values = append(values, value)
	}

	// 准备更新的字段
	var updates []string
	if len(updateFields) == 0 {
		// 如果没有指定更新字段，更新除主键外的所有字段
		for _, field := range fields {
			if field != m.primaryKey {
				updates = append(updates, fmt.Sprintf("%s=VALUES(%s)", field, field))
			}
		}
	} else {
		// 只更新指定的字段
		for _, field := range updateFields {
			if _, exists := record[field]; exists {
				updates = append(updates, fmt.Sprintf("%s=VALUES(%s)", field, field))
			}
		}
	}

	// 构建 SQL 语句
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) ON DUPLICATE KEY UPDATE %s",
		m.table,
		strings.Join(fields, ", "),
		strings.Repeat("?, ", len(fields)-1)+"?",
		strings.Join(updates, ", "),
	)

	kv = append(kv, "sql", query, "args", values)

	// 执行 SQL
	result, err := exec(m.client, query, values...)
	if err != nil {
		return nil, 0, err
	}

	affected, err = result.RowsAffected()
	if err != nil {
		return nil, 0, err
	}

	// 获取插入或更新后的数据
	var whereCondition []Option
	if pkValue, ok := record[m.primaryKey]; ok {
		whereCondition = append(whereCondition, WhereEq(m.primaryKey, pkValue))
	} else {
		// 如果没有主键，使用所有字段作为条件
		var conditions []Option
		for field, value := range record {
			conditions = append(conditions, WhereEq(field, value))
		}
		whereCondition = append(whereCondition, conditions...)
	}

	row := m.SelectOne(whereCondition...)
	if row.Err != nil {
		return nil, affected, row.Err
	}

	return Record(row.Data), affected, nil
}

func (m *model) Delete(opt ...Option) (ok bool, err error) {
	if len(opt) == 0 {
		return false, errors.New("danger, delete query must with some condition")
	}

	if m.err != nil {
		return false, m.err
	}

	opt = append(opt, table(m.table))
	if m.fakeDelKey != "" {
		m.enableValidator = false
		defer func() {
			m.enableValidator = true
		}()
		return m.Update(map[string]interface{}{m.fakeDelKey: 1}, opt...)
	}

	var kv []interface{}
	defer dbLog(m.ctx, "Delete", time.Now(), &err, &kv)

	_sql, args := DeleteBuilder(opt...)
	kv = append(kv, "slq", _sql, "args", args)

	result, err := exec(m.client, _sql, args...)
	if err != nil {
		return false, err
	}
	effect, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return effect > int64(0), nil
}

func (m *model) Exec(query string, args ...interface{}) (sql.Result, error) {
	return m.client.Exec(query, args...)
}

func (m *model) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return m.client.Query(query, args...)
}

func (m *model) hookInput(record map[string]interface{}) (map[string]interface{}, error) {
	for k, v := range m.columnHook {
		for field, val := range record {
			if k == field {
				overVal, err := v.Input(record, val)
				if err != nil {
					return nil, err
				}
				record[field] = overVal
			}
		}
	}
	return record, nil
}

func (m *model) recordToKV(record map[string]interface{}) (ks []string, vs []interface{}) {
	for k, v := range record {
		ks = append(ks, k)
		vs = append(vs, v)
	}

	return ks, vs
}
