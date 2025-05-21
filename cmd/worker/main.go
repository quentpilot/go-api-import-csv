package main

import (
	"go-csv-import/internal/logger"
	"go-csv-import/internal/queue"
)

func main() {
	if err := logger.InitCurrent("worker", false); err != nil {
		panic(err)
	}
	logger.Current.Info("Worker is listening...")
	queue.ConsumeImportJobs()
}
