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
		UseDb:      true,
	})
	self.Services = container.LoadConsumerServices(self.Conf)
	self.WatchForReload()

	s := server.New(&self.Conf.Http)

	s.LoadRoutes(server.UploadRouter{
		HttpConfig: &self.Conf.Http,
		AmqpConfig: &self.Conf.Amqp,
		Services:   self.Services,
	})

	s.Run()
}
