package queue

import (
	"go-csv-import/internal/config"
	"log/slog"

	"github.com/streadway/amqp"
)

type AmqpQueueHandler interface {
	getChannel() (*amqp.Connection, *amqp.Channel, error)
	Publish() error
	Consume() <-chan amqp.Delivery
}

type AmqpQueue struct {
	Config *config.ApmqConfig
}

func NewAmqpQueue(c *config.ApmqConfig) *AmqpQueue {
	return &AmqpQueue{Config: c}
}

func (q *AmqpQueue) getChannel(queue string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(q.Config.Dsn)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}

	_, err = ch.QueueDeclare(queue, true, false, false, false, nil)
	return conn, ch, err
}

func (q *AmqpQueue) Publish(queue string, body []byte) error {
	conn, ch, err := q.getChannel(queue)
	if err != nil {
		return err
	}
	defer conn.Close()
	defer ch.Close()

	return ch.Publish("", queue, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

func (q *AmqpQueue) Consume(queue string, autoAck bool) <-chan amqp.Delivery {
	_, ch, err := q.getChannel(queue)
	if err != nil {
		slog.Error("Connect to RabbitMQ", "error", err)
		panic(err)
	}

	msgs, err := ch.Consume(queue, "", autoAck, false, false, false, nil)
	if err != nil {
		slog.Error("Error while consuming message", "error", err)
		panic(err)
	}

	return msgs
}
