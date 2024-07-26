package xdb

import (
	"time"

	"github.com/spf13/cast"

	"github.com/daodao97/xgo/xlog"
)

func Info(msg string, kv ...interface{}) {
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

func Error(msg string, kv ...interface{}) {
	xlog.Error(msg, kv...)
}

func dbLog(prefix string, start time.Time, err *error, kv *[]interface{}) {
	tc := time.Since(start)

	_log := []any{
		xlog.String("method", prefix),
		xlog.String("scope", "xdb"),
		xlog.Any("duration", tc),
	}

	for i := 0; i < len(*kv); i++ {
		if i%2 == 0 {
			key := (*kv)[i]
			val := key
			if indexExists(*kv, i+1) {
				val = (*kv)[i+1]
			}
			_log = append(_log, xlog.Any(cast.ToString(key), val))
		}
	}

	if *err != nil {
		_log = append(_log, xlog.Any("error", *err))
		xlog.Error("query", _log...)
		return
	}
	xlog.Debug("query", _log...)
}

func indexExists(arr []any, index int) bool {
	return index >= 0 && index < len(arr)
}
