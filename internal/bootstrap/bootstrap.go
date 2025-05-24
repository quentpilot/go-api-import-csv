package bootstrap

import (
	"go-csv-import/internal/app"
	"go-csv-import/internal/config"
	"go-csv-import/internal/logger"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var appOnce sync.Once

func Init(c app.AppConfig) *app.Application {
	appOnce.Do(func() {
		initEnvConfig(&c)

		l := iniLogger(c)

		a := &app.Application{
			Config: &c,
		}

		a.SetLogger(l)

		app.Set(a)
	})

	return app.Get()
}

func iniLogger(c app.AppConfig) *slog.Logger {
	if c.LoggerName == "" {
		c.LoggerName = "root"
	}

	l, err := logger.InitCurrent(c.LoggerName, c.Logger.Level, false)
	if err != nil {
		panic(err)
	}

	return l
}

func initEnvConfig(c *app.AppConfig) {
	c.Logger.Load()
	c.Http.Load()
	c.Amqp.Load()
}

// WatchForReload listen SIGHUP, reload .env and update app configuration.
func WatchForReload() {
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGHUP)

		for range sigChan {
			app.Log().Info("Reload configuration (SIGHUP)")

			if err := config.ReloadEnv(); err != nil {
				slog.Error("Failed to reload .env", "error", err)
				continue
			}

			app.Config().Http.Load()
			app.Config().Logger.Load()

			newLogger, err := logger.InitCurrent(app.Config().LoggerName, app.Config().Logger.Level, false)
			if err != nil {
				slog.Error("Failed to reload logger", "error", err)
			} else {
				app.Get().SetLogger(newLogger)
				app.Log().Info("Configuration reloaded", "level", app.Config().Logger.Level)
				app.Get().PrintConfig()
				//fmt.Printf("app.Logger ptr: %p\n", app.Get().Logger())
				//fmt.Printf("slog.Default() ptr: %p\n", slog.Default())
			}
		}
	}()
}
