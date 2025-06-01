package amqp

import "github.com/streadway/amqp"

func (q *AmqpQueue) Publish(message AmqpMessage, tag string) error {
	conn, ch, err := q.connect()
	if err != nil {
		return err
	}
	defer conn.Close()
	defer ch.Close()

	return ch.Publish("", q.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        message.Get(),
		Type:        tag,
	})
}
