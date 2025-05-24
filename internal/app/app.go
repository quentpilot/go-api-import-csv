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
	a.Logger().Debug(fmt.Sprintf("%#v", a.Conf.LoggerName))
	a.Logger().Debug(fmt.Sprintf("%#v", a.Conf.Logger))
	a.Logger().Debug(fmt.Sprintf("%#v", a.Conf.Http))
	a.Logger().Debug(fmt.Sprintf("%#v", a.Conf.Amqp))
	a.Logger().Debug(fmt.Sprintf("%#v", a.Conf.Db))
	a.Logger().Debug(fmt.Sprintf("%#v", a.Conf.UseDb))
}

func (a *Application) Logger() *slog.Logger {
	return a.logger.Load()
}

func (a *Application) SetLogger(l *slog.Logger) {
	a.logger.Store(l)
}

func (a *Application) Log() *slog.Logger {
	return a.Logger()
}

func (a *Application) Config() *config.AppConfig {
	return a.Conf
}

func (a *Application) LoadConfig() {
	a.Conf.Logger.Load()
	a.Conf.Http.Load()
	a.Conf.Amqp.Load()
	a.Conf.Db.Load()
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

			a.LoadConfig()

			newLogger, err := logger.InitCurrent(a.Conf.LoggerName, a.Conf.Logger.Level, false)
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
