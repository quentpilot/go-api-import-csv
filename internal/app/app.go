package app

import (
	"fmt"
	"go-csv-import/internal/config"
	"go-csv-import/internal/logger"
	"log/slog"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
)

var application *Application

// Application holds the application modules like app env config, logger, etc.
type Application struct {
	logger atomic.Pointer[slog.Logger]
	Conf   *AppConfig
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
	a.Logger().Debug(fmt.Sprintf("%#v", a.Config().LoggerName))
	a.Logger().Debug(fmt.Sprintf("%#v", a.Config().Logger))
	a.Logger().Debug(fmt.Sprintf("%#v", a.Config().Http))
	a.Logger().Debug(fmt.Sprintf("%#v", a.Config().Amqp))
	a.Logger().Debug(fmt.Sprintf("%#v", a.Config().Db))
	a.Logger().Debug(fmt.Sprintf("%#v", a.Config().UseDb))
}

func (a *Application) Logger() *slog.Logger {
	return a.logger.Load()
}

func (a *Application) SetLogger(l *slog.Logger) {
	a.logger.Store(l)
}

func (a *Application) Log() *slog.Logger {
	return Get().Logger()
}

func (a *Application) Config() *AppConfig {
	return Get().Conf
}

func (a *Application) HttpConfig() config.HttpConfig {
	return a.Config().Http
}

func (a *Application) AmqpConfig() config.ApmqConfig {
	return a.Config().Amqp
}

func (a *Application) DbConfig() config.DbConfig {
	return a.Config().Db
}

// WatchForReload listen SIGHUP, reload .env and update app configuration.
func (a *Application) WatchForReload() {
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGHUP)

		for range sigChan {
			slog.Info("Reload configuration (SIGHUP)")

			if err := config.ReloadEnv(); err != nil {
				slog.Error("Failed to reload .env", "error", err)
				continue
			}

			a.Config().Http.Load()
			a.Config().Logger.Load()

			newLogger, err := logger.InitCurrent(a.Config().LoggerName, a.Config().Logger.Level, false)
			if err != nil {
				slog.Error("Failed to reload logger", "error", err)
			} else {
				a.SetLogger(newLogger)
				a.Log().Info("Configuration reloaded", "level", a.Config().Logger.Level)
				a.PrintConfig()
				//fmt.Printf("app.Logger ptr: %p\n", app.Get().Logger())
				//fmt.Printf("slog.Default() ptr: %p\n", slog.Default())
			}
		}
	}()
}
