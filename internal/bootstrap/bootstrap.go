package bootstrap

import (
	"go-csv-import/internal/app"
	"go-csv-import/internal/logger"
	"log/slog"
	"sync"
)

var appOnce sync.Once

func Init(c app.Config) *app.Application {
	appOnce.Do(func() {
		l := iniLogger(c)

		a := &app.Application{
			Logger: l,
		}

		app.Set(a)
	})

	return app.Get()
}

func iniLogger(c app.Config) *slog.Logger {
	if c.LoggerName == "" {
		c.LoggerName = "root"
	}

	if err := logger.InitCurrent(c.LoggerName, false); err != nil {
		panic(err)
	}

	return logger.Current
}
