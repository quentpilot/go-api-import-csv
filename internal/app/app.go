package app

import (
	"fmt"
	"go-csv-import/internal/config"
	"log/slog"
	"sync/atomic"
)

var application *Application

// Application holds the application modules like app env config, logger, etc.
type Application struct {
	logger atomic.Pointer[slog.Logger]
	Config *AppConfig
}

// Config holds the application modules parameters when initializing the application.
type AppConfig struct {
	LoggerName string              // File name for the current logger (default: "root")
	Logger     config.LoggerConfig // Logger configuration
	Http       config.HttpConfig   // HTTP server configuration
	Amqp       config.ApmqConfig   // AMQP server configuration
	Db         config.DbConfig     // Database server configuration
	UseDb      bool                // Whether to open a database connection (default: false)
}

func Set(a *Application) {
	application = a

	a.PrintConfig()
}

func Get() *Application {
	if application == nil {
		panic("Application not initialized. Make sure to call bootstrap.Init() before using the application.")
	}

	return application
}

func (a *Application) PrintConfig() {
	a.Logger().Debug(fmt.Sprintf("%#v", application.Config.LoggerName))
	a.Logger().Debug(fmt.Sprintf("%#v", application.Config.Logger))
	a.Logger().Debug(fmt.Sprintf("%#v", application.Config.Http))
	a.Logger().Debug(fmt.Sprintf("%#v", application.Config.Amqp))
	a.Logger().Debug(fmt.Sprintf("%#v", application.Config.Db))
	a.Logger().Debug(fmt.Sprintf("%#v", application.Config.UseDb))
}

func (a *Application) Logger() *slog.Logger {
	return a.logger.Load()
}

func (a *Application) SetLogger(l *slog.Logger) {
	a.logger.Store(l)
}

func Log() *slog.Logger {
	return Get().Logger()
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

func DbConfig() config.DbConfig {
	return Get().Config.Db
}
