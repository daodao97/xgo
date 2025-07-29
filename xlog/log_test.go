package xlog

import (
	"errors"
	"testing"
)

func TestLog(t *testing.T) {
	SetLogger(StdoutTextPretty())
	Debug("test", String("key", "value"), Map("map", map[string]any{"key": "value"}), Err(errors.New("test")))
}
