package bootstrap

import (
	"go-csv-import/internal/app"
	"go-csv-import/internal/config"
	"go-csv-import/internal/db"
	"go-csv-import/internal/logger"
	"log/slog"
)

func Load(c *config.AppConfig) *app.Application {
	initEnvConfig(c)
	l := iniLogger(c)

	a := &app.Application{
		Conf: c,
	}

	a.SetLogger(l)

	initDatabase(c)

	a.PrintConfig()

	//app.Get().Services = InitServices()

	return a
}

// Creates a new slogger
func iniLogger(c *config.AppConfig) *slog.Logger {
	if c.LoggerName == "" {
		c.LoggerName = "root"
	}

	l, err := logger.InitCurrent(c.LoggerName, c.Logger.Level, false)
	if err != nil {
		panic(err)
	}

	return l
}

// Loads all environment variable to through "config" package
func initEnvConfig(c *config.AppConfig) {
	c.Logger.Load()
	c.Http.Load()
	c.Amqp.Load()
	c.Db.Load()
}

// Opens a new database connection
func initDatabase(c *config.AppConfig) {
	if c.UseDb {
		if err := db.Connect(&c.Db); err != nil {
			slog.Error("Failed to connect to database", "error", err)
			panic(err)
		}
		db.AutoMigrate()
	}
}
