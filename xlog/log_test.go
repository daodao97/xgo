package xlog

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
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

func TestCallerInfoWithDifferentHandlers(t *testing.T) {
	tests := []struct {
		name   string
		logger func() *slog.Logger
	}{
		{
			name: "StdoutJson",
			logger: func() *slog.Logger {
				var buf bytes.Buffer
				return slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
					Level:     slog.LevelDebug,
					AddSource: true,
				}))
			},
		},
		{
			name: "StdoutText",
			logger: func() *slog.Logger {
				var buf bytes.Buffer
				return slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
					Level:     slog.LevelDebug,
					AddSource: true,
				}))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			var handler slog.Handler

			if tt.name == "StdoutJson" {
				handler = slog.NewJSONHandler(&buf, &slog.HandlerOptions{
					Level:     slog.LevelDebug,
					AddSource: true,
				})
			} else {
				handler = slog.NewTextHandler(&buf, &slog.HandlerOptions{
					Level:     slog.LevelDebug,
					AddSource: true,
				})
			}

			logger := slog.New(handler)
			// 直接调用 logger.Info 以确保 caller 信息指向测试文件
			logger.Info("test message", String("key", "value"))

			output := strings.TrimSpace(buf.String())

			// 验证输出包含文件信息
			if !strings.Contains(output, "log_test.go") {
				t.Errorf("Expected output to contain 'log_test.go', got: %s", output)
			}

			// 对于 JSON handler，验证 source 字段
			if tt.name == "StdoutJson" {
				var logData map[string]any
				if err := json.Unmarshal([]byte(output), &logData); err != nil {
					t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
				}

				source, ok := logData["source"]
				if !ok {
					t.Fatalf("Expected 'source' field in JSON output\nOutput: %s", output)
				}

				sourceMap, ok := source.(map[string]any)
				if !ok {
					t.Fatalf("Expected 'source' to be a map, got %T\nOutput: %s", source, output)
				}
				if _, ok := sourceMap["file"]; !ok {
					t.Errorf("Expected 'file' in source map\nOutput: %s", output)
				}
				if _, ok := sourceMap["line"]; !ok {
					t.Errorf("Expected 'line' in source map\nOutput: %s", output)
				}
			}
		})
	}
}

func TestCallerInfoWithWrapperFunction(t *testing.T) {
	// 测试通过包装函数调用时，caller 信息指向包装函数所在的文件
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}))
	SetLogger(logger)

	// 通过包装函数调用，caller 应该指向 log.go
	Info("test caller info with wrapper", String("key", "value"))

	// 解析 JSON 输出
	var logData map[string]any
	output := strings.TrimSpace(buf.String())
	if err := json.Unmarshal([]byte(output), &logData); err != nil {
		t.Fatalf("Failed to parse JSON log output: %v\nOutput: %s", err, output)
	}

	// 验证 source 字段存在
	source, ok := logData["source"]
	if !ok {
		t.Fatalf("Expected 'source' field in log output, but not found\nOutput: %s", output)
	}

	sourceMap, ok := source.(map[string]any)
	if !ok {
		t.Fatalf("Expected 'source' to be a map, got %T\nOutput: %s", source, output)
	}

	// 验证文件路径（通过包装函数调用时，应该指向 log.go）
	file, ok := sourceMap["file"].(string)
	if !ok {
		t.Fatalf("Expected 'file' field in source, got %T\nOutput: %s", sourceMap["file"], output)
	}

	// 通过包装函数调用时，caller 应该指向 log.go
	if !strings.Contains(file, "log.go") {
		t.Errorf("Expected file path to contain 'log.go' when using wrapper function, got: %s", file)
	}

	// 验证行号存在
	line, ok := sourceMap["line"].(float64)
	if !ok {
		t.Fatalf("Expected 'line' field in source to be a number, got %T\nOutput: %s", sourceMap["line"], output)
	}

	lineNum := int(line)
	if lineNum <= 0 {
		t.Errorf("Expected line number to be positive, got: %d", lineNum)
	}

	t.Logf("Caller info with wrapper verified: file=%s, line=%d", file, lineNum)
}

func TestDefaultOptionsIncludeCaller(t *testing.T) {
	// 测试 NewOptions 默认包含 AddSource
	opts := NewOptions()
	if !opts.AddSource {
		t.Error("Expected NewOptions() to have AddSource=true by default")
	}

	// 测试可以禁用 AddSource
	opts = NewOptions(WithAddSource(false))
	if opts.AddSource {
		t.Error("Expected WithAddSource(false) to disable AddSource")
	}
}
