package amqp

import "github.com/streadway/amqp"

/*
Publish sends message to AMQP queue

	message: the body content to send
	tag: the AMQP message type to handle
*/
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
