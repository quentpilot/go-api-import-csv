package logger

import (
	"context"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// NewDefault initialize the default logger file name "./logs/root.log".
// It overrides the default slog logger with a new one.
func NewDefault(level string, useJSON bool) (*slog.Logger, error) {
	l, err := New("root", level, useJSON)
	if err != nil {
		return nil, err
	}

	return l, nil
}

// NewCurrent initialize the wanted logger file name "./logs/<name>.log"
// It overrides the default slog logger with a new one.
func NewCurrent(name string, level string, useJSON bool) (*slog.Logger, error) {
	l, err := New(name, level, useJSON)
	if err != nil {
		return nil, err
	}

	return l, nil
}

// New returns a dedicated logger with a separate file. It write both to STDOUT and a file.
// It overrides the default slog logger with a new one.
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

	options.Level = ConvLogLevel(level)

	if useJSON {
		handler = slog.NewJSONHandler(output, &options)
	} else {
		//handler = slog.NewTextHandler(output, &options)
		handler = newTextHandler(output, &options)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger, nil
}

// newTextHandler creates a new slog.TextHandler with custom level names.
func newTextHandler(output io.Writer, opts *slog.HandlerOptions) *slog.TextHandler {
	levelNames := map[slog.Level]string{
		slog.LevelDebug:        "DEBUG",
		slog.LevelInfo:         "INFO",
		slog.LevelWarn:         "WARN",
		slog.LevelError:        "ERROR",
		ConvLogLevel("notice"): "NOTICE",
		ConvLogLevel("trace"):  "TRACE",
	}

	handler := slog.NewTextHandler(output, &slog.HandlerOptions{
		Level: opts.Level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				switch v := a.Value.Any().(type) {
				case slog.Level:
					if name, ok := levelNames[v]; ok {
						return slog.String(slog.LevelKey, name)
					}
				case string:
					// already a string, just return it
					return a
				default:
					// fallback : keep the original value
					return a
				}
			}
			return a
		}})

	return handler
}

func Trace(msg string, args ...any) {
	slog.Log(context.Background(), ConvLogLevel("trace"), msg, args...)
}

func Debug(msg string, args ...any) {
	slog.Log(context.Background(), ConvLogLevel("debug"), msg, args...)
}

func Notice(msg string, args ...any) {
	slog.Log(context.Background(), ConvLogLevel("notice"), msg, args...)
}

func Info(msg string, args ...any) {
	slog.Log(context.Background(), ConvLogLevel("info"), msg, args...)
}

func Warn(msg string, args ...any) {
	slog.Log(context.Background(), ConvLogLevel("warn"), msg, args...)
}

func Error(msg string, args ...any) {
	slog.Log(context.Background(), ConvLogLevel("error"), msg, args...)
}

func Fatal(msg string, args ...any) {
	slog.Log(context.Background(), ConvLogLevel("fatal"), msg, args...)
}

// Converts a string log level to slog.Level integer value
func ConvLogLevel(level string) slog.Level {
	level = strings.ToLower(level)

	switch level {
	case "trace":
		return slog.Level(-8)
	case "debug":
		return slog.LevelDebug
	case "notice":
		return slog.Level(2)
	case "warning", "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "fatal":
		return slog.Level(10)
	default:
		return slog.LevelInfo
	}
}
