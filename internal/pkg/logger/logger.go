package logger

import (
	"io"
	"log/slog"
	"os"
)

const logFileName = "app.log"

// logWriter открывает файл app.log и возвращает writer в файл + stderr (и в файл, и в консоль).
// При ошибке открытия файла возвращает только stderr.
func logWriter() io.Writer {
	f, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return os.Stderr
	}
	return io.MultiWriter(f, os.Stderr)
}

// New возвращает логгер с текстовым выводом в файл app.log в корне проекта и уровнем Info.
func New() *slog.Logger {
	return slog.New(slog.NewTextHandler(logWriter(), &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// NewWithLevel возвращает логгер с заданным уровнем (debug, info, warn, error).
func NewWithLevel(level string) *slog.Logger {
	var l slog.Level
	switch level {
	case "debug":
		l = slog.LevelDebug
	case "info":
		l = slog.LevelInfo
	case "warn", "warning":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}
	return slog.New(slog.NewTextHandler(logWriter(), &slog.HandlerOptions{
		Level: l,
	}))
}
