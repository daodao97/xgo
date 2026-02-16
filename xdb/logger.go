package xdb

import (
	"context"
	"log/slog"
	"reflect"
	"time"

	"github.com/spf13/cast"

	"github.com/daodao97/xgo/xlog"
)

func Info(msg string, kv ...any) {
	var _log []any
	for i := 0; i < len(kv); i++ {
		if i%2 == 0 {
			key := (kv)[i]
			val := (kv)[i+1]
			_log = append(_log, xlog.Any(cast.ToString(key), val))
		}
	}
	xlog.Debug(msg, _log...)
}

func Error(msg string, kv ...any) {
	xlog.Error(msg, kv...)
}

func dbLog(ctx context.Context, prefix string, start time.Time, err *error, kv *[]any) {
	tc := time.Since(start)

	_log := []any{
		xlog.String("method", prefix),
		xlog.String("scope", "xdb"),
		xlog.Any("duration", tc),
	}

	var sqlStmt string
	var rawArgs any
	var hasSQL bool

	for i := 0; i < len(*kv); i++ {
		if i%2 == 0 {
			key := (*kv)[i]
			val := key
			if indexExists(*kv, i+1) {
				val = (*kv)[i+1]
			}
			_log = append(_log, xlog.Any(cast.ToString(key), val))

			keyStr := cast.ToString(key)
			if !hasSQL && keyStr == "sql" {
				if stmt, ok := val.(string); ok {
					sqlStmt = stmt
					hasSQL = true
				}
			}
			if rawArgs == nil && keyStr == "args" {
				rawArgs = val
			}
		}
	}

	if hasSQL {
		if fullSQL := buildFullSQL(sqlStmt, rawArgs); fullSQL != "" {
			_log = append(_log, xlog.String("full_sql", fullSQL))
		}
		_log = removeLogFields(_log, "sql", "args")
	}

	if *err != nil {
		_log = append(_log, xlog.Any("error", *err))
		xlog.ErrorC(ctx, "query", _log...)
		return
	}
	xlog.DebugC(ctx, "query", _log...)
}

func indexExists(arr []any, index int) bool {
	return index >= 0 && index < len(arr)
}

func removeLogFields(logs []any, keys ...string) []any {
	if len(logs) == 0 || len(keys) == 0 {
		return logs
	}

	keySet := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		keySet[key] = struct{}{}
	}

	filtered := logs[:0]
	for _, item := range logs {
		if attr, ok := item.(slog.Attr); ok {
			if _, exists := keySet[attr.Key]; exists {
				continue
			}
		}
		filtered = append(filtered, item)
	}

	return filtered
}

func buildFullSQL(sqlStmt string, args any) (full string) {
	if sqlStmt == "" {
		return ""
	}

	argsSlice, ok := toAnySlice(args)
	if !ok {
		if args != nil {
			return ""
		}
	}

	defer func() {
		if r := recover(); r != nil {
			full = ""
		}
	}()

	return renderSQL(sqlStmt, argsSlice)
}

func toAnySlice(val any) ([]any, bool) {
	if val == nil {
		return nil, true
	}

	switch v := val.(type) {
	case []any:
		return v, true
	}

	rv := reflect.ValueOf(val)
	if !rv.IsValid() {
		return nil, false
	}

	k := rv.Kind()
	if k != reflect.Slice && k != reflect.Array {
		return nil, false
	}

	length := rv.Len()
	result := make([]any, length)
	for i := 0; i < length; i++ {
		result[i] = rv.Index(i).Interface()
	}

	return result, true
}
