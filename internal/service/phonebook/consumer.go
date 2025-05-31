package phonebook

import (
	"context"
	"fmt"
	"go-csv-import/internal/amqp"
	"go-csv-import/internal/config"
	"go-csv-import/internal/handlers/worker"
	"go-csv-import/internal/logger"
	"time"
)

// NewPhonebookConsumer creates a new instance of PhonebookHandler for consuming messages from the AMQP queue.
// As we need all the configuration (AMQP, HTTP, and DB), it accepts all three configurations as parameters.
func NewPhonebookConsumer(a *config.ApmqConfig, h *config.HttpConfig, d *config.DbConfig, s *worker.MessageProgressStore) *PhonebookHandler {
	self := NewPhonebookPublisher(a, h)

	self.ProgressStore = s
	self.Uploader = NewContactUploader(h, d, s)

	return self
}

// Consume listens for messages from the AMQP queue and processes them.
func (p *PhonebookHandler) Consume(ctx context.Context) {

	for msg := range p.Queue.Consume(true) {
		ctxT, cancel := context.WithTimeout(ctx, p.AmqpConfig.Lifetime)
		defer cancel()
		logger.Trace("Message received from queue", "message", msg.Body)

		// Decode the message body into a FileMessage struct.
		var file *FileMessage
		message := amqp.NewJsonMessageDecoder(msg.Body)
		err := message.Decode(&file)
		if err != nil {
			logger.Error("Decode AMQP message", "body", msg.Body, "error", err, "type", fmt.Sprintf("%T", err))
			continue
		}

		start := time.Now()
		logger.Info("Treating file", "file", file.FilePath)

		if err := p.Uploader.Upload(ctxT, file); err != nil {
			p.ProgressStore.SetError(file.Uuid, err)
			p.printTypedErrors(err, file)
		} else {
			logger.Info("File successful treated", "file", file.FilePath, "time", time.Since(start))
		}

		file.Remove()
		logger.Trace("Message acknowledged")

		select {
		case <-ctxT.Done():
			if ctxT.Err() != nil {
				switch ctxT.Err() {
				case context.Canceled:
					logger.Warn("Worker cancelled")
					return
				case context.DeadlineExceeded:
					logger.Warn("Deadline exceeded")
				default:
					logger.Warn("Unknown cancellation reason")
					return
				}
			}
		default:
		}
	}
}
