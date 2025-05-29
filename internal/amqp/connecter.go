package amqp

import (
	"github.com/streadway/amqp"
)

// AmqpMessageHandler represents the way to transport AMQP messages between producer input and consumer output.
type AmqpQueueHandler interface {
	connect() (*amqp.Connection, *amqp.Channel, error) // internal server connection
	Publish(message AmqpMessage) error                 // Publish a message to the RabbitMQ queue
	Consume() <-chan amqp.Delivery                     // Get messages from queue. No error handling here, as it panics on error
}

// AmqpQueue implements the AmqpQueueHandler interface and represents a RabbitMQ queue with its connection details.
type AmqpQueue struct {
	Dsn        string
	Name       string
	connection *amqp.Connection
	channel    *amqp.Channel
}

// NewAmqpQueue creates a new instance of AmqpQueue with the provided DSN and queue name.
func NewAmqpQueue(dns string, queueName string) *AmqpQueue {
	return &AmqpQueue{
		Dsn:  dns,
		Name: queueName,
	}
}

// connect establishes a connection to the RabbitMQ server and declares the queue.
func (q *AmqpQueue) connect() (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(q.Dsn)
	if err != nil {
		return nil, nil, err
	}
	q.connection = conn

	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}
	q.channel = ch

	_, err = ch.QueueDeclare(q.Name, true, false, false, false, nil)
	return conn, ch, err
}

func (q *AmqpQueue) Close() error {
	if q.channel != nil {
		return q.channel.Close()
	}
	if q.connection != nil {
		return q.connection.Close()
	}
	return nil
}
