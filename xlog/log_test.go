package xlog

import (
	"errors"
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	SetLogger(StdoutTextPretty())
	Debug("test",
		Int("int", 1),
		String("key", "value"),
		Map("map", map[string]any{"key": "value"}),
		Err(errors.New("test")),
		Time("time", time.Now()),
		Duration("duration", time.Second),
		Bool("bool", true),
		Float64("float64", 1.23),
		Any("any", map[string]any{"key": "value"}),
		Any("any", []any{1, "2", true}),
		Any("any", []byte("test")),
	)
}
