package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/TechTeam-ZUS/zus-go-common/config"
	"github.com/TechTeam-ZUS/zus-go-common/notify"
)

var logger *slog.Logger

type ctxKey string

const loggerCtxKey ctxKey = "logger"

var levels = map[string]slog.Level{
	"Debug": slog.LevelDebug,
	"Info":  slog.LevelInfo,
	"Warn":  slog.LevelWarn,
	"Error": slog.LevelError,
}

func Init() *slog.Logger {
	cfg := config.LoadLogger()

	handler := getHandler(cfg.HandlerType, parseLevel(cfg.LogLevel))

	logger = slog.New(handler).With(
		"service", cfg.ServiceName,
	)

	return logger
}

func GetLogger() *slog.Logger {
	if logger == nil {
		Init()
	}

	return logger
}

// WithContext stores a logger in the context.
func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey, l)
}

// FromContext retrieves the logger from context, falling back to the global logger.
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerCtxKey).(*slog.Logger); ok {
		return l
	}
	return GetLogger()
}

func Info(msg string, args ...any)  { GetLogger().Info(msg, args...) }
func Debug(msg string, args ...any) { GetLogger().Debug(msg, args...) }

func Warn(msg string, args ...any) {
	GetLogger().Warn(msg, args...)
}

func Error(msg string, args ...any) {
	GetLogger().Error(msg, args...)
}

func Fatal(msg string, args ...any) {
	GetLogger().Error(msg, args...)
	os.Exit(1)
}

// Infof, Debugf, Warnf, Errorf, and Fatalf are printf-style variants: format
// is resolved via fmt.Sprintf before logging, for callers building a message
// dynamically instead of attaching structured key-value fields.
func Infof(format string, args ...any)  { GetLogger().Info(fmt.Sprintf(format, args...)) }
func Debugf(format string, args ...any) { GetLogger().Debug(fmt.Sprintf(format, args...)) }

func Warnf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	GetLogger().Warn(msg)
}

func Errorf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	GetLogger().Error(msg)
}

func Fatalf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	GetLogger().Error(msg)
	os.Exit(1)
}

func WarnfWithNotifier(title, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	GetLogger().Warn(msg)
	notify.GetBot().Notify(slog.LevelWarn, title, msg)
}

func ErrorfWithNotifier(title, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	GetLogger().Error(msg)
	notify.GetBot().Notify(slog.LevelError, title, msg)
}

func parseLevel(level string) slog.Level {
	l, ok := levels[level]
	if !ok {
		return slog.LevelDebug
	}

	return l
}

func getHandler(handlerType string, logLevel slog.Level) slog.Handler {
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	if handlerType == "json" {
		return slog.NewJSONHandler(os.Stdout, opts)
	}

	return newTextHandler(os.Stdout, logLevel)
}
