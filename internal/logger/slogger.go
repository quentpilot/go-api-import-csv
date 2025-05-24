package logger

import (
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// InitDefault initialize the default logger file name "./logs/root.log"
func InitDefault(level string, useJSON bool) (*slog.Logger, error) {
	l, err := New("root", level, useJSON)
	if err != nil {
		return nil, err
	}

	return l, nil
}

// InitDefault initialize the wanted logger file name "./logs/<name>.log"
func InitCurrent(name string, level string, useJSON bool) (*slog.Logger, error) {
	l, err := New(name, level, useJSON)
	if err != nil {
		return nil, err
	}

	return l, nil
}

// New returns a dedicated logger with a separate file
func New(name string, level string, useJSON bool) (*slog.Logger, error) {
	log.Println("New slog handlers:", name, level)
	logsDir := "logs"
	if strings.HasSuffix(os.Args[0], ".test") { // avoid creating logs/ dir in package when test mode
		logsDir = os.TempDir()
	}

	if err := os.MkdirAll(logsDir, os.ModePerm); err != nil {
		return nil, err
	}

	logPath := filepath.Join(logsDir, name+".log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	output := io.MultiWriter(os.Stdout, f)

	var handler slog.Handler
	var options slog.HandlerOptions

	options.Level = convLogLevel(level)

	if useJSON {
		handler = slog.NewJSONHandler(output, &options)
	} else {
		handler = slog.NewTextHandler(output, &options)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger, nil
}

// Converts a string log level to slog.Level integer value
func convLogLevel(level string) slog.Level {
	level = strings.ToLower(level)

	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
