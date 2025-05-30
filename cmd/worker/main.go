package main

import (
	"context"
	"go-csv-import/internal/bootstrap"
	"go-csv-import/internal/config"
	"go-csv-import/internal/container"
	"go-csv-import/internal/logger"
	"os"
	"os/signal"
)

func main() {
	self := bootstrap.Load(&config.AppConfig{
		LoggerName: "worker",
		UseDb:      true,
	})
	self.Services = container.LoadConsumerServices(self.Conf)
	self.WatchForReload()
	logger.Info("Phonebook worker is listening...")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	logger.Trace("Signal context created", "signal", "Interrupt")

	self.Services.PhonebookUploader.Consume(ctx)

	self.Services.PhonebookUploader.Close()

	logger.Info("...Shutdown Phonebook Worker")
}
