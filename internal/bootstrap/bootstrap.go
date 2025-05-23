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
		initEnvConfig(&c)

		l := iniLogger(c)

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
