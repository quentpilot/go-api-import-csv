package main

import (
	"context"
	"go-csv-import/internal/bootstrap"
	"go-csv-import/internal/config"
	"go-csv-import/internal/container"
	"os"
	"os/signal"
)

func main() {
	self := bootstrap.Load(&config.AppConfig{
		LoggerName: "worker",
		UseDb:      true,
	})
	self.Services = container.LoadServices(self.Conf)
	self.WatchForReload()
	self.Log().Info("Phonebook worker is listening...")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	self.Services.PhonebookUploader.Consume(ctx)

	self.Services.PhonebookUploader.Close()

	self.Log().Info("...Shutdown Phonebook Worker")
}
