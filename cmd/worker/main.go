package main

import (
	"go-csv-import/internal/app"
	"go-csv-import/internal/bootstrap"
	"go-csv-import/internal/queue"
)

func main() {
	bootstrap.Init(app.AppConfig{
		LoggerName: "worker",
	})

	app.Logger().Info("Worker is listening...")
	queue.ConsumeImportJobs()
}
