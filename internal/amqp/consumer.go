package amqp

import (
	"log/slog"

	"github.com/streadway/amqp"
)

func (q *AmqpQueue) Consume(autoAck bool) <-chan amqp.Delivery {
	_, ch, err := q.connect()
	if err != nil {
		slog.Error("Cannot connect to AMQP server", "error", err)
		panic(err)
	}

	msgs, err := ch.Consume(q.Name, "", autoAck, false, false, false, nil)
	if err != nil {
		slog.Error("Cannot consuming AMQP message queue", "error", err)
		panic(err)
	}

	return msgs
}
