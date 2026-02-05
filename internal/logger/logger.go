package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
)

var (
	once   sync.Once
	logger *slog.Logger
)

const (
	reset   = "\033[0m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	gray    = "\033[90m"
	white   = "\033[97m"
)

type PrettyHandler struct {
	w     io.Writer
	level slog.Leveler
	mu    sync.Mutex
}

func NewPrettyHandler(w io.Writer, level slog.Level) *PrettyHandler {
	return &PrettyHandler{w: w, level: level}
}

func (h *PrettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	timestamp := r.Time.Format("15:04:05")

	var levelColor, levelText string
	switch r.Level {
	case slog.LevelDebug:
		levelColor = gray
		levelText = "DBG"
	case slog.LevelInfo:
		levelColor = green
		levelText = "INF"
	case slog.LevelWarn:
		levelColor = yellow
		levelText = "WRN"
	case slog.LevelError:
		levelColor = red
		levelText = "ERR"
	}

	fmt.Fprintf(h.w, "%s%s%s %s%-3s%s %s",
		gray, timestamp, reset,
		levelColor, levelText, reset,
		r.Message,
	)

	r.Attrs(func(a slog.Attr) bool {
		fmt.Fprintf(h.w, " %s%s%s=%v", cyan, a.Key, reset, a.Value)
		return true
	})

	fmt.Fprintln(h.w)
	return nil
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	return h
}

func Init() *slog.Logger {
	once.Do(func() {
		format := os.Getenv("LOG_FORMAT")
		levelStr := os.Getenv("LOG_LEVEL")

		level := slog.LevelInfo
		switch levelStr {
		case "debug":
			level = slog.LevelDebug
		case "warn":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		}

		var handler slog.Handler
		if format == "json" {
			handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
		} else {
			handler = NewPrettyHandler(os.Stdout, level)
		}

		logger = slog.New(handler)
		slog.SetDefault(logger)
	})
	return logger
}

func Get() *slog.Logger {
	if logger == nil {
		return Init()
	}
	return logger
}

func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}

func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}
