package xlog

import (
	"os"
	"testing"
)

func TestLog(t *testing.T) {
	SetLogger(NewHandler(os.Stdout, &ColorOptions{}))
	Debug("test", String("key", "value"), Map("map", map[string]any{"key": "value"}))
}
