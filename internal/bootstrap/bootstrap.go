package bootstrap

import (
	"go-csv-import/internal/app"
	"go-csv-import/internal/db"
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
			Conf: &c,
		}

		a.SetLogger(l)
		app.Set(a)

		initDatabase(&c)

		//app.Get().Services = InitServices()
	})

	return app.Get()
}

// Creates a new slogger
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

// Loads all environment variable to through "config" package
func initEnvConfig(c *app.AppConfig) {
	c.Logger.Load()
	c.Http.Load()
	c.Amqp.Load()
	c.Db.Load()
}

// Opens a new database connection
func initDatabase(c *app.AppConfig) {
	if c.UseDb {
		if err := db.Connect(c.Db); err != nil {
			slog.Error("Failed to connect to database", "error", err)
			panic(err)
		}
		db.AutoMigrate()
	}
}

/* func InitServices() *Services {
	return &Services{
		Queue: service.NewImportFileQueue(),
	}
} */
