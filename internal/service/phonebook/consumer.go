package phonebook

import (
	"context"
	"fmt"
	"go-csv-import/internal/amqp"
	"go-csv-import/internal/config"
	"log/slog"
	"time"
)

// NewPhonebookConsumer creates a new instance of PhonebookHandler for consuming messages from the AMQP queue.
// As we need all the configuration (AMQP, HTTP, and DB), it accepts all three configurations as parameters.
func NewPhonebookConsumer(a *config.ApmqConfig, h *config.HttpConfig, d *config.DbConfig) *PhonebookHandler {
	self := NewPhonebookPublisher(a, h)

	self.uploader = NewContactUploader(h, d)

	return self
}

// Consume listens for messages from the AMQP queue and processes them.
func (p *PhonebookHandler) Consume(ctx context.Context) {

	for msg := range p.Queue.Consume(true) {
		ctxT, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		slog.Debug("Message received from queue", "message", msg.Body)

		// Decode the message body into a FileMessage struct.
		var file *FileMessage
		message := amqp.NewJsonMessageDecoder(msg.Body)
		err := message.Decode(&file)
		if err != nil {
			slog.Error("Decode AMQP message", "body", msg.Body, "error", err, "type", fmt.Sprintf("%T", err))
			continue
		}

		start := time.Now()
		slog.Info("Treating file", "file", file.FilePath)

		if err := p.uploader.Upload(ctxT, file); err != nil {
			p.printTypedErrors(err, file)
		} else {
			slog.Info("File successful treated", "file", file.FilePath, "time", time.Since(start))
		}

		file.Remove()
		slog.Info("Message acknowledged")
		select {
		case <-ctxT.Done():
			if ctxT.Err() != nil {
				switch ctxT.Err() {
				case context.Canceled:
					slog.Warn("Worker cancelled")
					return
				case context.DeadlineExceeded:
					slog.Warn("Deadline exceeded")
				default:
					slog.Warn("Unknown cancellation reason")
					return
				}
			}
		default:
		}
	}
}
