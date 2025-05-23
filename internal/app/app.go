package app

import (
	"log/slog"
)

var app *Application

// Application holds the application modules like app env config, logger, etc.
type Application struct {
	Logger *slog.Logger
}

// Config holds the application modules parameters when initializing the application.
type Config struct {
	LoggerName string // File name for the current logger (default: "root")
}

func Set(a *Application) {
	app = a
}

func Get() *Application {
	if app == nil {
		panic("application not initialized")
	}
	return app
}

func Logger() *slog.Logger {
	return Get().Logger
}
