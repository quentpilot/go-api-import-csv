package main

import (
	"go-csv-import/internal/bootstrap"
	"go-csv-import/internal/config"
	"go-csv-import/internal/container"
)

func main() {
	self := bootstrap.Load(&config.AppConfig{
		LoggerName: "worker",
		UseDb:      true,
	})
	self.Services = container.LoadServices(self.Conf)
	self.WatchForReload()
	self.Log().Info("Worker is listening...")

	self.Services.ImportFileQueue.Consume()

	/* worker := service.NewImportFileQueueConsumer(self.AmqpConfig(), self.HttpConfig(), self.DbConfig())

	worker.Consume() */
	self.Log().Info("...Shutdown Worker")
}
