package xlog

import (
	"context"
	"log/slog"
	"os"
	"sync/atomic"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/daodao97/xgo/xtrace"
)

func init() {
	SetLogger(StdoutTextPretty())

	// 默认 extractor：request_id（从 ctx string-key 读，无则跳过）
	RegisterCtxExtractor(func(ctx context.Context) []slog.Attr {
		if v, ok := ctx.Value("request_id").(string); ok && v != "" {
			return []slog.Attr{slog.String("request_id", v)}
		}
		return nil
	})

	// 默认 extractor：traceid（xtrace 包内部维护 ctx key）
	RegisterCtxExtractor(func(ctx context.Context) []slog.Attr {
		if tid := xtrace.FromTraceId(ctx); tid != "" {
			return []slog.Attr{slog.String("traceid", tid)}
		}
		return nil
	})
}

var opts = slog.HandlerOptions{
	Level:       slog.LevelDebug,
	AddSource:   true,
	ReplaceAttr: replaceSourceAttr,
}

var logger = StdoutTextPretty()

// CtxExtractor 从 ctx 中提取一组 slog.Attr。无值时返回 nil（或空切片）。
// 通过 RegisterCtxExtractor 注册后，所有带 ctx 的日志调用都会自动应用。
type CtxExtractor func(ctx context.Context) []slog.Attr

var ctxExtractors atomic.Pointer[[]CtxExtractor]

// RegisterCtxExtractor 注册一个 Context Attr 提取器。
// 通常在 init() 或 main() 早期调用。运行时调用也安全（lock-free copy-on-write）。
// 传入 nil 会被忽略。
func RegisterCtxExtractor(ex CtxExtractor) {
	if ex == nil {
		return
	}
	for {
		old := ctxExtractors.Load()
		var oldSlice []CtxExtractor
		if old != nil {
			oldSlice = *old
		}
		newSlice := make([]CtxExtractor, 0, len(oldSlice)+1)
		newSlice = append(newSlice, oldSlice...)
		newSlice = append(newSlice, ex)
		if ctxExtractors.CompareAndSwap(old, &newSlice) {
			return
		}
	}
}

// withCtxAttrs 遍历所有已注册的 extractor，把它们返回的 attr 追加到 args。
func withCtxAttrs(ctx context.Context, args ...any) []any {
	if ctx == nil {
		return args
	}
	p := ctxExtractors.Load()
	if p == nil {
		return args
	}
	for _, ex := range *p {
		for _, attr := range ex(ctx) {
			args = append(args, attr)
		}
	}
	return args
}

func SetLogger(l *slog.Logger) {
	logger = l
	slog.SetDefault(l)
}

func GetLogger() *slog.Logger {
	return logger
}

func StdoutText(opts ...Option) *slog.Logger {
	_opts := NewOptions(opts...)
	return slog.New(slog.NewTextHandler(os.Stdout, &_opts.HandlerOptions))
}

func StdoutTextPretty(opts ...Option) *slog.Logger {
	_opts := NewOptions(opts...)

	return slog.New(NewPrettyHandler(os.Stdout, PrettyHandlerOptions{
		SlogOpts: _opts.HandlerOptions,
	}))
}

func StdoutJson(opts ...Option) *slog.Logger {
	_opts := NewOptions(opts...)

	return slog.New(slog.NewJSONHandler(os.Stdout, &_opts.HandlerOptions))
}

func FileJson(fileName string) *slog.Logger {
	r := &lumberjack.Logger{
		Filename:   fileName,
		LocalTime:  true,
		MaxSize:    1,
		MaxAge:     3,
		MaxBackups: 5,
		Compress:   true,
	}
	return slog.New(slog.NewJSONHandler(r, &opts))
}

func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

func Error(msg string, args ...any) {
	logger.Error(msg, args...)
}

func Warn(msg string, args ...any) {
	logger.Warn(msg, args...)
}

func DebugCtx(ctx context.Context, msg string, args ...any) {
	logger.DebugContext(ctx, msg, withCtxAttrs(ctx, args...)...)
}

func InfoCtx(ctx context.Context, msg string, args ...any) {
	logger.InfoContext(ctx, msg, withCtxAttrs(ctx, args...)...)
}

func ErrorCtx(ctx context.Context, msg string, args ...any) {
	logger.ErrorContext(ctx, msg, withCtxAttrs(ctx, args...)...)
}

func WarnCtx(ctx context.Context, msg string, args ...any) {
	logger.WarnContext(ctx, msg, withCtxAttrs(ctx, args...)...)
}

func err(err error) slog.Attr {
	return slog.Any("err", err)
}

var (
	String   = slog.String
	Int      = slog.Int
	Int64    = slog.Int64
	Uint64   = slog.Uint64
	Float64  = slog.Float64
	Bool     = slog.Bool
	Time     = slog.Time
	Duration = slog.Duration
	Any      = slog.Any
	Group    = slog.Group
	Err      = err

	DebugC = DebugCtx
	InfoC  = InfoCtx
	WarnC  = WarnCtx
	ErrorC = ErrorCtx
)

func Map(key string, value any) slog.Attr {
	return slog.Attr{Key: key, Value: slog.AnyValue(value)}
}

func AnySlice(key string, value []any) slog.Attr {
	return slog.Attr{Key: key, Value: slog.AnyValue(value)}
}
