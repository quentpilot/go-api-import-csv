package app

import (
	"fmt"
	"go-csv-import/internal/config"
	"log/slog"
)

var app *Application

// Application holds the application modules like app env config, logger, etc.
type Application struct {
	Logger *slog.Logger
	Config *AppConfig
}

// Config holds the application modules parameters when initializing the application.
type AppConfig struct {
	LoggerName string              // File name for the current logger (default: "root")
	Logger     config.LoggerConfig // Logger configuration
	Http       config.HttpConfig   // HTTP server configuration
	Amqp       config.ApmqConfig   // AMQP server configuration
}

func Set(a *Application) {
	app = a

	printConfig()
}

func Get() *Application {
	if app == nil {
		panic("Application not initialized. Make sure to call bootstrap.Init() before using the application.")
	}
	return app
}

func printConfig() {
	Logger().Debug(fmt.Sprintf("%#v", app.Config.Logger))
	Logger().Debug(fmt.Sprintf("%#v", app.Config.Http))
	Logger().Debug(fmt.Sprintf("%#v", app.Config.Amqp))
}

func Logger() *slog.Logger {
	return Get().Logger
}

func Config() *AppConfig {
	return Get().Config
}

func HttpConfig() config.HttpConfig {
	return Get().Config.Http
}

func AmqpConfig() config.ApmqConfig {
	return Get().Config.Amqp
}
