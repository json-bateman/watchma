package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"
)

// ANSI color codes, bolded
const (
	ColorReset  = "\033[0m"
	ColorBlack  = "\033[1;30m"
	ColorRed    = "\033[1;31m"
	ColorYellow = "\033[1;33m"
	ColorGreen  = "\033[1;32m"
	ColorBlue   = "\033[1;34m"
	ColorGray   = "\033[1;90m"
)

type ColorHandler struct {
	level slog.Level
}

func New(level slog.Level) *slog.Logger {
	return slog.New(&ColorHandler{level: level})
}

func (h *ColorHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *ColorHandler) Handle(_ context.Context, r slog.Record) error {
	var color string

	switch {
	case r.Level >= slog.LevelError:
		color = ColorRed
	case r.Level >= slog.LevelWarn:
		color = ColorYellow
	case r.Level >= slog.LevelInfo:
		color = ColorGreen
	default:
		color = ColorGray
	}

	timestamp := r.Time.Format(time.RFC3339)
	msg := r.Message

	fmt.Fprintf(os.Stderr, "%s[%s] [%s] %s%s\n", color, timestamp, r.Level.String(), msg, ColorReset)

	if r.NumAttrs() > 0 {
		r.Attrs(func(a slog.Attr) bool {
			fmt.Fprintf(os.Stderr, "%s    - %s: %v%s\n", color, a.Key, a.Value, ColorReset)
			return true
		})
	}

	return nil
}

// Need these to satisfy slog's interface, even though they don't do anything
func (h *ColorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *ColorHandler) WithGroup(name string) slog.Handler {
	return h
}
