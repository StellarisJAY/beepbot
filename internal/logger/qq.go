package logger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/StellarisJAY/beepbot/internal/config"
	"github.com/tencent-connect/botgo"
)

type qqChannelLogger struct {
	logger *slog.Logger
}

// Debug implements [log.Logger].
func (q *qqChannelLogger) Debug(v ...interface{}) {
	q.logger.Debug(fmt.Sprint(v...))
}

// Debugf implements [log.Logger].
func (q *qqChannelLogger) Debugf(format string, v ...interface{}) {
	q.logger.Debug(fmt.Sprintf(format, v...))
}

// Error implements [log.Logger].
func (q *qqChannelLogger) Error(v ...interface{}) {
	q.logger.Error(fmt.Sprint(v...))
}

// Errorf implements [log.Logger].
func (q *qqChannelLogger) Errorf(format string, v ...interface{}) {
	q.logger.Error(fmt.Sprintf(format, v...))
}

// Info implements [log.Logger].
func (q *qqChannelLogger) Info(v ...interface{}) {
	q.logger.Info(fmt.Sprint(v...))
}

// Infof implements [log.Logger].
func (q *qqChannelLogger) Infof(format string, v ...interface{}) {
	q.logger.Info(fmt.Sprintf(format, v...))
}

// Sync implements [log.Logger].
func (q *qqChannelLogger) Sync() error {
	// slog.Logger doesn't require explicit sync
	return nil
}

// Warn implements [log.Logger].
func (q *qqChannelLogger) Warn(v ...interface{}) {
	q.logger.Warn(fmt.Sprint(v...))
}

// Warnf implements [log.Logger].
func (q *qqChannelLogger) Warnf(format string, v ...interface{}) {
	q.logger.Warn(fmt.Sprintf(format, v...))
}

func initQQChannelLogger(config config.Logging) error {
	loggingConfig := config.QQ
	path := strings.TrimSpace(loggingConfig.File)
	if path == "" {
		path = "./logs/qq_channel.log"
	}
	level := strings.ToLower(strings.TrimSpace(loggingConfig.Level))

	absPath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	dir := filepath.Dir(absPath)
	os.MkdirAll(dir, os.ModeDir)
	file, err := os.OpenFile(absPath, os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}

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
	logger := slog.New(slog.NewTextHandler(file, &slog.HandlerOptions{Level: logLevel}))
	qqLogger := &qqChannelLogger{logger: logger}
	botgo.SetLogger(qqLogger)
	return nil
}
