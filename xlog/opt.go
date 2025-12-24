package xlog

import (
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"time"
)

type Options struct {
	TimeFormat  string
	SrcFileMode SourceFileMode
	Trace       bool
	slog.HandlerOptions
}

type Option = func(opts *Options)

// replaceSourceAttr 替换 source 属性，跳过 xlog 包中的包装函数
// 这个方法通过遍历当前调用栈，找到第一个不在 xlog 包中的调用位置
func replaceSourceAttr(groups []string, a slog.Attr) slog.Attr {
	// 只处理 source 属性
	if a.Key != slog.SourceKey {
		return a
	}

	// 从调用栈中找到第一个不在 xlog 包中的调用
	// 需要跳过：runtime.Caller, replaceSourceAttr, handler 内部调用, xlog 包装函数
	// 通常需要跳过 5-8 层
	for skip := 5; skip < 20; skip++ {
		pc, file, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		fnName := fn.Name()

		// 跳过 xlog 包中的函数和 slog 内部的函数
		if strings.Contains(fnName, "/xlog.") ||
			strings.Contains(fnName, "\\xlog.") ||
			strings.Contains(fnName, "log/slog") {
			continue
		}

		// 找到第一个不在 xlog 包中的调用，创建新的 source 属性
		return slog.Group(slog.SourceKey,
			slog.String("caller", fmt.Sprintf("%s:%d", file, line)),
		)
	}

	// 如果找不到，返回原始属性（fallback）
	return a
}

func NewOptions(opts ...Option) *Options {
	o := &Options{
		TimeFormat:  time.DateTime,
		SrcFileMode: ShortFile,
		HandlerOptions: slog.HandlerOptions{
			Level:       slog.LevelDebug,
			AddSource:   true,
			ReplaceAttr: replaceSourceAttr,
		},
	}

	for _, opt := range opts {
		opt(o)
	}

	// 如果用户自定义了 ReplaceAttr，需要合并处理
	// 保存用户自定义的 ReplaceAttr
	userReplaceAttr := o.ReplaceAttr
	o.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
		// 先应用我们的 source 替换
		a = replaceSourceAttr(groups, a)
		// 如果用户也定义了 ReplaceAttr，再应用用户的替换
		if userReplaceAttr != nil {
			a = userReplaceAttr(groups, a)
		}
		return a
	}

	return o
}

func WithLevel(level slog.Leveler) Option {
	return func(opts *Options) {
		opts.Level = level
	}
}

func WithAddSource(addSource bool) Option {
	return func(opts *Options) {
		opts.AddSource = addSource
	}
}

func WithReplaceAttr(fn func(groups []string, a slog.Attr) slog.Attr) Option {
	return func(opts *Options) {
		opts.ReplaceAttr = fn
	}
}

func WithTimeFormat(format string) Option {
	return func(opts *Options) {
		opts.TimeFormat = format
	}
}

func WithSrcFileMode(mode SourceFileMode) Option {
	return func(opts *Options) {
		opts.SrcFileMode = mode
	}
}

func WithTrace(trace bool) Option {
	return func(opts *Options) {
		opts.Trace = trace
	}
}
