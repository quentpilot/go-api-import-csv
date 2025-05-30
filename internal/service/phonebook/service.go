package phonebook

import (
	"context"
	"go-csv-import/internal/amqp"
	"go-csv-import/internal/config"
	"go-csv-import/internal/db"
	"go-csv-import/internal/logger"
	"log/slog"
)

type PhonebookService interface {
	Publish(message FileMessage) error
	Consume(ctx context.Context)
}

type PhonebookHandler struct {
	AmqpConfig *config.ApmqConfig
	HttpConfig *config.HttpConfig
	Queue      *amqp.AmqpQueue
	Uploader   *ContactUploader
}

// Close closes the AMQP queue and database connection.
func (p *PhonebookHandler) Close() error {
	err := db.Close()
	if err != nil {
		slog.Warn("Error closing database connection", "error", err)
	}

	if p.Queue != nil {
		err = p.Queue.Close()
		if err != nil {
			slog.Warn("Error closing AMQP queue", "error", err)
		}
	}

	logger.Trace("Phonebook service closed", "error", err)

	return err
}
