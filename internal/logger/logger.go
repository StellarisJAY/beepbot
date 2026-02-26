package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/StellarisJAY/beepbot/internal/config"
)

func InitLogger(config config.Logging) {
	logFilePath := config.File
	logFilePath = strings.TrimSpace(logFilePath)
	if logFilePath == "" {
		logFilePath = "./logs/beepbot.log"
	}
	absPath, err := filepath.Abs(logFilePath)
	if err != nil {
		panic(err)
	}
	dir := filepath.Dir(absPath)
	os.MkdirAll(dir, os.ModeDir)
	file, err := os.OpenFile(absPath, os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}

	writer := io.MultiWriter(file, os.Stdout)

	level := strings.ToLower(config.Level)
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	format := strings.ToLower(config.Format)
	var handler slog.Handler
	switch format {
	case "json":
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: logLevel})
	case "text":
		handler = slog.NewTextHandler(writer, &slog.HandlerOptions{Level: logLevel})
	default:
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: logLevel})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}
