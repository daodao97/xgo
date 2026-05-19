package xlog

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/daodao97/xgo/xtrace"
)

// snapshotExtractors 在测试开始时保存当前注册表，测试结束后恢复。
// 用于隔离测试间的全局状态污染。
func snapshotExtractors(t *testing.T) {
	t.Helper()
	saved := ctxExtractors.Load()
	t.Cleanup(func() {
		ctxExtractors.Store(saved)
	})
}

func TestRegisterCtxExtractor_appendsToRegistry(t *testing.T) {
	snapshotExtractors(t)
	// 清空，便于断言长度
	ctxExtractors.Store(nil)

	ex1 := func(ctx context.Context) []slog.Attr { return nil }
	ex2 := func(ctx context.Context) []slog.Attr { return nil }

	RegisterCtxExtractor(ex1)
	p := ctxExtractors.Load()
	if p == nil || len(*p) != 1 {
		t.Fatalf("expected len 1 after first register, got %v", p)
	}

	RegisterCtxExtractor(ex2)
	p = ctxExtractors.Load()
	if p == nil || len(*p) != 2 {
		t.Fatalf("expected len 2 after second register, got %v", p)
	}
}

func TestRegisterCtxExtractor_ignoresNil(t *testing.T) {
	snapshotExtractors(t)
	ctxExtractors.Store(nil)

	RegisterCtxExtractor(nil)
	p := ctxExtractors.Load()
	if p != nil && len(*p) != 0 {
		t.Fatalf("expected empty/nil registry after registering nil, got %v", p)
	}
}

func TestWithCtxAttrs_nilCtxReturnsArgsUnchanged(t *testing.T) {
	snapshotExtractors(t)
	ctxExtractors.Store(nil)

	args := []any{slog.String("k", "v")}
	var nilCtx context.Context // intentional: verify withCtxAttrs is nil-ctx safe
	got := withCtxAttrs(nilCtx, args...)
	if len(got) != 1 {
		t.Fatalf("expected len 1, got %d", len(got))
	}
}

func TestWithCtxAttrs_appliesAllExtractors(t *testing.T) {
	snapshotExtractors(t)
	ctxExtractors.Store(nil)

	RegisterCtxExtractor(func(ctx context.Context) []slog.Attr {
		return []slog.Attr{slog.String("from_ex1", "v1")}
	})
	RegisterCtxExtractor(func(ctx context.Context) []slog.Attr {
		return []slog.Attr{
			slog.String("from_ex2_a", "va"),
			slog.String("from_ex2_b", "vb"),
		}
	})

	got := withCtxAttrs(context.Background(), slog.String("user", "manual"))
	if len(got) != 4 {
		t.Fatalf("expected 4 args (1 manual + 1 + 2), got %d: %#v", len(got), got)
	}
}

func TestWithCtxAttrs_nilReturnSkipped(t *testing.T) {
	snapshotExtractors(t)
	ctxExtractors.Store(nil)

	RegisterCtxExtractor(func(ctx context.Context) []slog.Attr { return nil })

	got := withCtxAttrs(context.Background(), slog.String("k", "v"))
	if len(got) != 1 {
		t.Fatalf("expected 1 arg (nil extractor adds nothing), got %d", len(got))
	}
}

// captureCtxLog 用 JSON handler 捕获一次 InfoCtx 输出，返回解析后的 map。
func captureCtxLog(t *testing.T, ctx context.Context, msg string, args ...any) map[string]any {
	t.Helper()
	var buf bytes.Buffer
	prev := GetLogger()
	t.Cleanup(func() { SetLogger(prev) })
	SetLogger(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	InfoCtx(ctx, msg, args...)

	var out map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &out); err != nil {
		t.Fatalf("parse log output: %v\nraw: %s", err, buf.String())
	}
	return out
}

func TestDefaultExtractor_RequestId(t *testing.T) {
	snapshotExtractors(t)
	// 不清空，依赖 init() 注册的默认 extractor

	ctx := context.WithValue(context.Background(), "request_id", "rid-123")
	out := captureCtxLog(t, ctx, "hello")

	if out["request_id"] != "rid-123" {
		t.Errorf("expected request_id=rid-123, got %v", out["request_id"])
	}
}

func TestDefaultExtractor_RequestIdMissingDoesNotPanic(t *testing.T) {
	snapshotExtractors(t)
	ctx := context.Background()
	out := captureCtxLog(t, ctx, "hello")
	if _, ok := out["request_id"]; ok {
		t.Errorf("expected no request_id field when ctx has none, got %v", out["request_id"])
	}
}

func TestDefaultExtractor_TraceId(t *testing.T) {
	snapshotExtractors(t)

	// xtrace.SetTraceId 是写入 trace id 的官方入口（xtrace/xtrace.go:80）。
	ctx := xtrace.SetTraceId(context.Background(), "trace-abc")
	out := captureCtxLog(t, ctx, "hello")

	if out["traceid"] != "trace-abc" {
		t.Errorf("expected traceid=trace-abc, got %v", out["traceid"])
	}
}

func TestInfoCtx_AppliesCustomExtractor(t *testing.T) {
	snapshotExtractors(t)

	RegisterCtxExtractor(func(ctx context.Context) []slog.Attr {
		if v, ok := ctx.Value("user_id").(string); ok {
			return []slog.Attr{slog.String("user_id", v)}
		}
		return nil
	})

	ctx := context.WithValue(context.Background(), "user_id", "u-42")
	out := captureCtxLog(t, ctx, "hello", slog.String("order_id", "o-1"))

	if out["user_id"] != "u-42" {
		t.Errorf("expected user_id=u-42, got %v", out["user_id"])
	}
	if out["order_id"] != "o-1" {
		t.Errorf("expected manual order_id=o-1 preserved, got %v", out["order_id"])
	}
}
