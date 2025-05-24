package main

import (
	"go-csv-import/internal/bootstrap"
	"go-csv-import/internal/config"
	"go-csv-import/internal/container"
	"go-csv-import/internal/server"
)

func main() {
	self := bootstrap.Load(&config.AppConfig{
		LoggerName: "api",
	})
	self.Services = container.LoadServices(self.Conf)
	self.WatchForReload()

	s := server.New()

	s.LoadRoutes(server.UploadRouter{
		HttpConfig: &self.Conf.Http,
		AmqpConfig: &self.Conf.Amqp,
	})

	s.Run(self.HttpConfig().Port)
}
