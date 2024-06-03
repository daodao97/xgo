package xlog

import (
	"log/slog"
	"time"
)

type Options struct {
	TimeFormat  string
	SrcFileMode SourceFileMode
	Trace       bool
	slog.HandlerOptions
}

type Option = func(opts *Options)

func NewOptions(opts ...Option) *Options {
	o := &Options{
		TimeFormat:  time.DateTime,
		SrcFileMode: ShortFile,
		HandlerOptions: slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	for _, opt := range opts {
		opt(o)
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
