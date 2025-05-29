package phonebook

import (
	"context"
	"go-csv-import/internal/amqp"
	"go-csv-import/internal/config"
)

type PhonebookService interface {
	Publish(message FileMessage) error
	Consume(ctx context.Context)
}

type PhonebookHandler struct {
	AmqpConfig *config.ApmqConfig
	HttpConfig *config.HttpConfig
	Queue      *amqp.AmqpQueue
	uploader   *ContactUploader
}
