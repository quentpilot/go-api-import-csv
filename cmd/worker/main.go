package main

import (
	"go-csv-import/internal/app"
	"go-csv-import/internal/bootstrap"
	"go-csv-import/internal/service"
)

func main() {
	self := bootstrap.Init(app.AppConfig{
		LoggerName: "worker",
		UseDb:      true,
	})
	self.WatchForReload()
	self.Log().Info("Worker is listening...")

	worker := service.NewImportFileQueueConsumer(self.AmqpConfig(), self.HttpConfig(), self.DbConfig())

	worker.Consume()
	self.Log().Info("...Shutdown Worker")
}
