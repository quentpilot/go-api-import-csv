package bootstrap

import (
	"go-csv-import/internal/app"
	"go-csv-import/internal/logger"
	"log/slog"
	"sync"
)

var appOnce sync.Once

func Init(c app.AppConfig) *app.Application {
	appOnce.Do(func() {
		l := iniLogger(c)
		initEnvConfig(&c)

		a := &app.Application{
			Logger: l,
			Config: &c,
		}

		app.Set(a)
	})

	return app.Get()
}

func iniLogger(c app.AppConfig) *slog.Logger {
	if c.LoggerName == "" {
		c.LoggerName = "root"
	}

	if err := logger.InitCurrent(c.LoggerName, false); err != nil {
		panic(err)
	}

	return logger.Current
}

func initEnvConfig(c *app.AppConfig) {
	c.Logger.Load()
	c.Http.Load()
	c.Amqp.Load()
}
