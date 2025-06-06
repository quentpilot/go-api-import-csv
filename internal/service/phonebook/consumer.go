package phonebook

import (
	"context"
	"go-csv-import/internal/config"
	"go-csv-import/internal/handlers/worker"
	"go-csv-import/internal/logger"
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
	msgHandler := p.NewMessageHandler()

	for msg := range p.Queue.Consume(false) {
		ctxT, cancel := context.WithTimeout(ctx, p.AmqpConfig.Lifetime)
		defer cancel()
		logger.Trace("Message received from queue", "type", msg.Type, "message", msg.Body)

		ack, err := msgHandler.Process(ctxT, msg)
		if err != nil {
			logger.Error("MessageHandler error", "tag", msg.Type, "error", err)
		}

		if ack {
			logger.Debug("Message acknowledged")
			msg.Ack(false)
		} else {
			logger.Debug("Message unacknowledged")
			msg.Nack(false, true)
		}

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
