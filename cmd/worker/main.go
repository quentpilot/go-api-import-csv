package main

import (
	"context"
	"fmt"
	"go-csv-import/internal/bootstrap"
	"go-csv-import/internal/config"
	"go-csv-import/internal/container"
	"go-csv-import/internal/handlers/worker"
	"go-csv-import/internal/logger"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	self := bootstrap.Load(&config.AppConfig{
		LoggerName: "worker",
		UseDb:      true,
	})
	progressStore := worker.NewMessageProgressStore()
	self.Services = container.LoadConsumerServices(self.Conf, progressStore)
	self.WatchForReload()
	logger.Info("Phonebook worker is listening...")

	// Run internal HTTP server for message progress
	go func() {
		if err := http.ListenAndServe(":9090", progressStore.Handler()); err != nil {
			logger.Error("Failed to start progress store server", "error", err)
			panic(fmt.Errorf("failed to start progress store server: %w", err))
		}
	}()

	// Run AMPQ consumer
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	logger.Trace("Signal context created", "signal", "Interrupt")

	self.Services.PhonebookUploader.Consume(ctx)

	self.Services.PhonebookUploader.Close()

	logger.Info("...Shutdown Phonebook Worker")
}
