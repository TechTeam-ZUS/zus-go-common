package common

import (
	"log/slog"
	"os"

	"github.com/TechTeam-ZUS/zus-go-common/config"
)

func Init() *slog.Logger {
	cfg := config.LoadLogger()

	handler := getHandler(cfg.HandlerType, parseLevel(cfg.LogLevel))

	return slog.New(handler).With(
		"service", cfg.ServiceName,
	)
}

func parseLevel(level string) slog.Level {
	switch level {
	case "Debug":
		return slog.LevelDebug
	case "Info":
		return slog.LevelInfo
	case "Warn":
		return slog.LevelWarn
	case "Error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func getHandler(handlerType string, logLevel slog.Level) slog.Handler {
	// if handlerType == "text"{
	return slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: logLevel,
		},
	)
	// }
	// todo: json handler
}
