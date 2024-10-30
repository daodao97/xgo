package xdb

import (
	"encoding/json"
	"time"

	"github.com/daodao97/xgo/xdb/interval/util"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
)

var ErrRowBindingType = errors.New("binding dest type must be *struct **struct")
var ErrRowsBindingType = errors.New("binding dest type must be *[]struct, *[]*struct")

type Record map[string]any

func (r Record) Binding(dest any) error {
	if !util.AllowType(dest, []string{"*struct", "**struct"}) {
		return ErrRowBindingType
	}

	return util.Binding(r, dest)
}

func (r Record) Get(key string) (any, bool) {
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
	if value, ok := v.(*time.Time); ok {
		return value
	}

	if value, ok := v.(time.Time); ok {
		return &value
	}

	return nil
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

func (r Record) GetDecimal(key string) *decimal.Decimal {
	v, ok := r[key]
	if !ok {
		return nil
	}

	if value, ok := v.(*decimal.Decimal); ok {
		return value
	}

	if value, ok := v.(decimal.Decimal); ok {
		return &value
	}

	return nil
}

// Deprecated: use Record instead
type Row struct {
	Data map[string]any
	Err  error
}

func (r *Row) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Data)
}

func (r *Row) Binding(dest any) error {
	if r.Err != nil {
		return r.Err
	}
	if !util.AllowType(dest, []string{"*struct", "**struct"}) {
		return ErrRowBindingType
	}

	return util.Binding(r.Data, dest)
}

func (r *Row) Get(key string) (any, bool) {
	v, ok := r.Data[key]
	return v, ok
}

func (r *Row) GetString(key string) string {
	v, ok := r.Data[key]
	if !ok {
		return ""
	}
	return cast.ToString(v)
}

func (r *Row) GetInt(key string) int {
	v, ok := r.Data[key]
	if !ok {
		return 0
	}
	return cast.ToInt(v)
}

func (r *Row) GetArray(key string) []any {
	v, ok := r.Data[key]
	if !ok {
		return []any{}
	}
	return cast.ToSlice(v)
}

func (r *Row) GetTime(key string) *time.Time {
	v, ok := r.Data[key]
	if !ok {
		return nil
	}

	if value, ok := v.(*time.Time); ok {
		return value
	}

	if value, ok := v.(time.Time); ok {
		return &value
	}

	return nil
}

func (r *Row) GetMap(key string) map[string]any {
	v, ok := r.Data[key]
	if !ok {
		return nil
	}
	return *v.(*map[string]any)
}

func (r *Row) GetAny(key string) any {
	v, ok := r.Data[key]
	if !ok {
		return nil
	}
	return v
}

type Rows struct {
	List []Row
	Err  error
}

func (r *Rows) Binding(dest any) error {
	if !util.AllowType(dest, []string{"*[]struct", "*[]*struct"}) {
		return ErrRowsBindingType
	}

	var source []map[string]any
	for _, v := range r.List {
		source = append(source, v.Data)
	}

	return util.Binding(source, dest)
}
