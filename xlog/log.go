package xlog

import (
	"context"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/daodao97/xgo/xtrace"
)

func init() {
	SetLogger(StdoutTextPretty())
}

var opts = slog.HandlerOptions{
	Level:       slog.LevelDebug,
	AddSource:   true,
	ReplaceAttr: replaceSourceAttr,
}

var logger = StdoutTextPretty()

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

func withTriceId(ctx context.Context, args ...any) []any {
	if ctx == nil {
		return args
	}
	if requestId := ctx.Value("request_id"); requestId != nil {
		args = append(args, String("request_id", requestId.(string)))
	}
	if triceId := xtrace.FromTraceId(ctx); triceId != "" {
		args = append(args, String("trice_id", triceId))
	}
	return args
}

func DebugCtx(ctx context.Context, msg string, args ...any) {
	logger.DebugContext(ctx, msg, withTriceId(ctx, args...)...)
}

func InfoCtx(ctx context.Context, msg string, args ...any) {
	logger.InfoContext(ctx, msg, withTriceId(ctx, args...)...)
}

func ErrorCtx(ctx context.Context, msg string, args ...any) {
	logger.ErrorContext(ctx, msg, withTriceId(ctx, args...)...)
}

func WarnCtx(ctx context.Context, msg string, args ...any) {
	logger.WarnContext(ctx, msg, withTriceId(ctx, args...)...)
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
