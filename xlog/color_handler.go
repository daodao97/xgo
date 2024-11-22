package xlog

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/fatih/color"
)

type SourceFileMode int

const (
	// Nop does nothing.
	Nop SourceFileMode = iota

	// ShortFile produces only the filename (for example main.go:69).
	ShortFile

	// LongFile produces the full file path (for example /home/frajer/go/src/myapp/main.go:69).
	LongFile
)

var DefaultOptions *ColorOptions = &ColorOptions{
	Level:       slog.LevelInfo,
	TimeFormat:  time.DateTime,
	SrcFileMode: ShortFile,
	Trace:       true,
}

type ColorOptions struct {
	// Level reports the minimum level to xlog.
	// Levels with lower levels are discarded.
	// If nil, the Handler uses [slog.LevelInfo].
	Level slog.Leveler

	// The time format.
	TimeFormat string

	// SrcFileMode is the source file mode.
	SrcFileMode SourceFileMode
	Trace       bool
}

type Handler struct {
	groups []string
	attrs  []slog.Attr

	opts ColorOptions

	mu  *sync.Mutex
	out io.Writer
}

// NewHandler creates a new Handler.
func NewHandler(out io.Writer, opts *ColorOptions) *Handler {
	h := &Handler{out: out, mu: &sync.Mutex{}}
	if opts != nil {
		h.opts = *opts
	}
	return h
}

func (h *Handler) clone() *Handler {
	return &Handler{
		groups: h.groups,
		attrs:  h.attrs,
		opts:   h.opts,
		mu:     h.mu,
		out:    h.out,
	}
}

// Enabled implements slog.Handler.Enabled .
func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

// Handle implements slog.Handler.Handle .
func (h *Handler) Handle(_ context.Context, r slog.Record) error {
	var bf bytes.Buffer

	if !r.Time.IsZero() {
		fmt.Fprint(&bf, color.New(color.Faint).Sprint(r.Time.Format(h.opts.TimeFormat)))
		fmt.Fprint(&bf, " ")
	}

	level := r.Level.String()
	switch r.Level {
	case slog.LevelDebug:
		level = color.MagentaString(level)
	case slog.LevelInfo:
		level = color.BlueString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}
	fmt.Fprint(&bf, level)

	fmt.Fprint(&bf, " ")

	if h.opts.Trace && r.PC != 0 {
		f, _ := runtime.CallersFrames([]uintptr{r.PC - 2}).Next()

		switch h.opts.SrcFileMode {
		case Nop:
			break
		case ShortFile:
			fmt.Fprintf(&bf, "%s:%d ", filepath.Base(f.File), f.Line)
		case LongFile:
			fmt.Fprintf(&bf, "%s:%d ", f.File, f.Line)
		}
	}

	//fmt.Fprint(&bf, color.HiWhiteString("| "))

	fmt.Fprint(&bf, r.Message)

	var attrs []slog.Attr
	attrs = append(attrs, h.attrs...)
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	for _, a := range attrs {
		fmt.Fprint(&bf, " ")
		for i, g := range h.groups {
			fmt.Fprint(&bf, color.New(color.FgCyan).Sprint(g))
			if i != len(h.groups) {
				fmt.Fprint(&bf, color.New(color.FgCyan).Sprint("."))
			}
		}

		value := a.Value.String()
		fmt.Fprint(&bf, color.New(color.FgCyan).Sprintf("%s=", a.Key)+value)
	}

	fmt.Fprint(&bf, "\n")

	h.mu.Lock()
	_, err := io.Copy(h.out, &bf)
	h.mu.Unlock()

	return err
}

// WithGroup implements slog.Handler.WithGroup .
func (h *Handler) WithGroup(name string) slog.Handler {
	h2 := h.clone()
	h2.groups = append(h2.groups, name)
	return h2
}

// WithAttrs implements slog.Handler.WithAttrs .
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h2 := h.clone()
	h2.attrs = append(h2.attrs, attrs...)
	return h2
}
