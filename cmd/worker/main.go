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

	/* if err := ctx.Err(); err != nil {
		switch err {
		case context.Canceled:
			slog.Warn("Worker annulé")
		case context.DeadlineExceeded:
			slog.Warn("Délai dépassé")
		default:
			slog.Warn("Annulation inconnue")
		}
	} */

	self.Log().Info("...Shutdown Worker")
}
