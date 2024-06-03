package xlog

import (
	"context"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

var opts = slog.HandlerOptions{
	Level: slog.LevelDebug,
}

var logger = StdoutText()

func SetLogger(l *slog.Logger) {
	logger = l
	slog.SetDefault(l)
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
	if requestId := ctx.Value("request_id"); requestId != nil {
		args = append([]any{requestId.(string)}, args...)
	}
	logger.DebugContext(ctx, msg, args...)
}

func InfoCtx(ctx context.Context, msg string, args ...any) {
	if requestId := ctx.Value("request_id"); requestId != nil {
		args = append([]any{requestId.(string)}, args...)
	}
	logger.InfoContext(ctx, msg, args...)
}

func ErrorCtx(ctx context.Context, msg string, args ...any) {
	if requestId := ctx.Value("request_id"); requestId != nil {
		args = append([]any{requestId.(string)}, args...)
	}
	logger.ErrorContext(ctx, msg, args...)
}

func WarnCtx(ctx context.Context, msg string, args ...any) {
	if requestId := ctx.Value("request_id"); requestId != nil {
		args = append([]any{requestId.(string)}, args...)
	}
	logger.WarnContext(ctx, msg, args...)
}

func Err(err error) slog.Attr {
	return slog.Any("err", err)
}
