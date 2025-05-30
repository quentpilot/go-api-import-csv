package app

import (
	"fmt"
	"go-csv-import/internal/config"
	"go-csv-import/internal/container"
	"go-csv-import/internal/logger"
	"log/slog"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
)

// Application holds the application modules like app env config, logger, etc.
type Application struct {
	logger   atomic.Pointer[slog.Logger]
	Conf     *config.AppConfig
	Services *container.Services
}

func (a *Application) PrintConfig() {
	logger.Debug(fmt.Sprintf("%#v", a.Conf.LoggerName))
	logger.Debug(fmt.Sprintf("%#v", a.Conf.Logger))
	logger.Debug(fmt.Sprintf("%#v", a.Conf.Http))
	logger.Debug(fmt.Sprintf("%#v", a.Conf.Amqp))
	logger.Debug(fmt.Sprintf("%#v", a.Conf.Db))
	logger.Debug(fmt.Sprintf("%#v", a.Conf.UseDb))
}

// Logger returns the logger for the application in a safe atomic pointer.
func (a *Application) Logger() *slog.Logger {
	return a.logger.Load()
}

// SetLogger sets the logger for the application in a safe atomic pointer.
func (a *Application) SetLogger(l *slog.Logger) {
	a.logger.Store(l)
}

// Log returns the singleton slogger instance of the application.
func (a *Application) Log() *slog.Logger {
	return a.Logger()
}

// Config returns the application configuration.
func (a *Application) Config() *config.AppConfig {
	return a.Conf
}

// LoadConfig loads the application configuration from environment variables.
func (a *Application) LoadConfig() {
	a.Conf.Logger.Load()
	a.Conf.Http.Load()
	a.Conf.Amqp.Load()
	a.Conf.Db.Load()
}

// HttpConfig returns the HTTP configuration of the application.
func (a *Application) HttpConfig() config.HttpConfig {
	return a.Config().Http
}

// AmqpConfig returns the AMQP configuration of the application.
func (a *Application) AmqpConfig() config.ApmqConfig {
	return a.Config().Amqp
}

// DbConfig returns the database configuration of the application.
func (a *Application) DbConfig() config.DbConfig {
	return a.Config().Db
}

// WatchForReload listen SIGHUP signal to reload .env file and update app configuration.
func (a *Application) WatchForReload() {
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGHUP)

		for range sigChan {
			logger.Trace("Reload configuration (SIGHUP)")

			if err := config.ReloadEnv(); err != nil {
				slog.Error("Failed to reload .env", "error", err)
				continue
			}

			a.LoadConfig()

			newLogger, err := logger.NewCurrent(a.Conf.LoggerName, a.Conf.Logger.Level, false)
			if err != nil {
				slog.Error("Failed to reload logger", "error", err)
			} else {
				a.SetLogger(newLogger)
				logger.Trace("Configuration reloaded", "level", a.Config().Logger.Level)
				a.PrintConfig()
			}
		}
	}()
}
