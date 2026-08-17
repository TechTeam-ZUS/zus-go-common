package logger

import (
	"log/slog"
	"os"

	"github.com/TechTeam-ZUS/zus-go-common/config"
)

var logger *slog.Logger

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
		logger = Init()
	}

	return logger
}

func Info(msg string, args ...any)  { GetLogger().Info(msg, args...) }
func Warn(msg string, args ...any)  { GetLogger().Warn(msg, args...) }
func Error(msg string, args ...any) { GetLogger().Error(msg, args...) }
func Debug(msg string, args ...any) { GetLogger().Debug(msg, args...) }

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

	return slog.NewTextHandler(os.Stdout, opts)
}
