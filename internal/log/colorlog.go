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
	ColorBlue   = "\033[34m" // WebSocket non bolded
	ColorGray   = "\033[1;90m"
)

const LevelWebSocket slog.Level = slog.Level(2)

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
	var levelLabel string

	switch {
	case r.Level == LevelWebSocket:
		color = ColorBlue
		levelLabel = "WS"
	case r.Level >= slog.LevelError:
		color = ColorRed
		levelLabel = "ERROR"
	case r.Level >= slog.LevelWarn:
		color = ColorYellow
		levelLabel = "WARN"
	case r.Level >= slog.LevelInfo:
		color = ColorGreen
		levelLabel = "INFO"
	default:
		color = ColorGray
		levelLabel = "OTHER"
	}

	timestamp := r.Time.Format(time.RFC3339)
	msg := r.Message

	fmt.Fprintf(os.Stderr, "%s[%s] [%s] %s%s\n", color, timestamp, levelLabel, msg, ColorReset)

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
