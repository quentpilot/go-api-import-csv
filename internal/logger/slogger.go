package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

var Default *slog.Logger
var Current *slog.Logger

// InitDefault initialize the default logger
func InitDefault(useJSON bool) error {
	l, err := New("root", useJSON)
	if err != nil {
		return err
	}
	Default = l
	return nil
}

func InitCurrent(name string, useJSON bool) error {
	l, err := New(name, useJSON)
	if err != nil {
		return err
	}
	Current = l
	return nil
}

// New returns a dedicated logger with a separate file
func New(name string, useJSON bool) (*slog.Logger, error) {
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
	if useJSON {
		handler = slog.NewJSONHandler(output, nil)
	} else {
		handler = slog.NewTextHandler(output, nil)
	}

	return slog.New(handler), nil
}
