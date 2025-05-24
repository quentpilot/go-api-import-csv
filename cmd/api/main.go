package main

import (
	"go-csv-import/internal/app"
	"go-csv-import/internal/bootstrap"
	"go-csv-import/internal/server"
)

func main() {
	self := bootstrap.Init(app.AppConfig{
		LoggerName: "api",
	})
	self.WatchForReload()

	s := server.New()

	s.LoadRoutes(server.UploadRouter{
		HttpConfig: self.HttpConfig(),
		AmqpConfig: self.AmqpConfig(),
	})

	s.Run(self.HttpConfig().Port)
}
