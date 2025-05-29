package phonebook

import (
	"go-csv-import/internal/amqp"
	"go-csv-import/internal/config"
)

// NewPhonebookPublisher creates a new instance of PhonebookHandler only for publishing messages to the AMQP queue.
// As we don't need to connect to the database, it set a minimal configuration with the AMQP and HTTP configurations.
func NewPhonebookPublisher(a *config.ApmqConfig, h *config.HttpConfig) *PhonebookHandler {
	self := &PhonebookHandler{
		AmqpConfig: a,
		HttpConfig: h,
	}

	self.Queue = amqp.NewAmqpQueue(a.Dsn, a.Queue)

	return self
}

// Publish sends a json string of FileMessage to the AMQP queue.
func (p *PhonebookHandler) Publish(message *FileMessage) error {
	body, err := amqp.NewJsonMessageEncoder(message)
	if err != nil {
		return err
	}

	return p.Queue.Publish(body)
}
