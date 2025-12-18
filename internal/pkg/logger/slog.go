package logger

import (
	"log/slog"
	"os"
)

type Log interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}

// 方便以后更换日志框架
type Logger struct {
	*slog.Logger
}

func NewLogger() *Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
		AddSource: true,
    })
	return &Logger{
		slog.New(handler),
	}
}

func (l *Logger) Info(msg string, args ...any) {
    l.Logger.Info(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.Logger.Error(msg, args...)
}