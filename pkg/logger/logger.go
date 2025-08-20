package logger

import (
	"log/slog"
	"os"
)

var (
	stdoutLogger *slog.Logger
)

func Init(lvl slog.Level) {
	stdoutHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	stdoutLogger = slog.New(stdoutHandler)
}

type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

func Debug(msg string, args ...interface{}) {
	stdoutLogger.Debug(msg, args...)
}

func Info(msg string, args ...interface{}) {
	stdoutLogger.Info(msg, args...)
}

func Warn(msg string, args ...interface{}) {
	stdoutLogger.Warn(msg, args...)
}

func Error(msg string, args ...interface{}) {
	stdoutLogger.Error(msg, args...)
}
