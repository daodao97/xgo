package xdb

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func renderSQL(query string, args []any) string {
	if len(args) == 0 || query == "" {
		return query
	}

	var builder strings.Builder
	argIndex := 0
	for i := 0; i < len(query); i++ {
		if query[i] == '?' && argIndex < len(args) {
			builder.WriteString(renderSQLArg(args[argIndex]))
			argIndex++
			continue
		}
		builder.WriteByte(query[i])
	}

	return builder.String()
}

func renderSQLArg(arg any) string {
	if arg == nil {
		return "NULL"
	}

	switch v := arg.(type) {
	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return fmt.Sprintf("%d", v)
	case float32:
		return formatFloat(float64(v))
	case float64:
		return formatFloat(v)
	case string:
		return quoteString(v)
	case []byte:
		return quoteString(string(v))
	case time.Time:
		return quoteString(v.Format(time.RFC3339Nano))
	case fmt.Stringer:
		return quoteString(v.String())
	case driver.Valuer:
		value, err := v.Value()
		if err != nil {
			return "NULL"
		}
		return renderSQLArg(value)
	}

	rv := reflect.ValueOf(arg)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return "NULL"
		}
		return renderSQLArg(rv.Elem().Interface())
	}

	return quoteString(fmt.Sprint(arg))
}

func quoteString(s string) string {
	escaped := strings.ReplaceAll(s, "'", "''")
	return "'" + escaped + "'"
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
